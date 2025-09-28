package filetree

import (
	"strings"
	"sync"

	"github.com/jesseduffield/lazygit/pkg/commands/models"
	"github.com/jesseduffield/lazygit/pkg/common"
	"github.com/jesseduffield/lazygit/pkg/gui/context/traits"
	"github.com/jesseduffield/lazygit/pkg/gui/types"
	"github.com/jesseduffield/lazygit/pkg/utils"
	"github.com/samber/lo"
)

type FileSectionKind int

const (
	FileSectionStaged FileSectionKind = iota
	FileSectionUnstaged
	FileSectionUntracked
)

type FileSection struct {
	Kind  FileSectionKind
	Nodes []*FileNode
}

var fileSectionOrder = []FileSectionKind{
	FileSectionStaged,
	FileSectionUnstaged,
	FileSectionUntracked,
}

type DisplayableTree interface {
	DisplayRoot() *Node[models.File]
}

type IFileTreeViewModel interface {
	IFileTree
	types.IListCursor
}

// This combines our FileTree struct with a cursor that retains information about
// which item is selected. It also contains logic for repositioning that cursor
// after the files are refreshed
type FileTreeViewModel struct {
	sync.RWMutex
	types.IListCursor
	IFileTree

	sectionsCache    []FileSection
	itemsCache       []*FileNode
	displayRootCache *Node[models.File]
}

var _ IFileTreeViewModel = &FileTreeViewModel{}

func NewFileTreeViewModel(getFiles func() []*models.File, common *common.Common, showTree bool) *FileTreeViewModel {
	fileTree := NewFileTree(getFiles, common, showTree)
	viewModel := &FileTreeViewModel{
		IFileTree: fileTree,
	}
	viewModel.IListCursor = traits.NewListCursor(func() int { return viewModel.Len() })
	return viewModel
}

func (self *FileTreeViewModel) GetSelected() *FileNode {
	if self.Len() == 0 {
		return nil
	}

	return self.Get(self.GetSelectedLineIdx())
}

func (self *FileTreeViewModel) GetSelectedItemId() string {
	item := self.GetSelected()
	if item == nil {
		return ""
	}

	return item.ID()
}

func (self *FileTreeViewModel) GetSelectedItems() ([]*FileNode, int, int) {
	if self.Len() == 0 {
		return nil, 0, 0
	}

	startIdx, endIdx := self.GetSelectionRange()

	nodes := []*FileNode{}
	for i := startIdx; i <= endIdx; i++ {
		nodes = append(nodes, self.Get(i))
	}

	return nodes, startIdx, endIdx
}

func (self *FileTreeViewModel) GetSelectedItemIds() ([]string, int, int) {
	selectedItems, startIdx, endIdx := self.GetSelectedItems()

	ids := lo.Map(selectedItems, func(item *FileNode, _ int) string {
		return item.ID()
	})

	return ids, startIdx, endIdx
}

func (self *FileTreeViewModel) GetSelectedFile() *models.File {
	node := self.GetSelected()
	if node == nil {
		return nil
	}

	return node.File
}

func (self *FileTreeViewModel) GetSelectedPath() string {
	node := self.GetSelected()
	if node == nil {
		return ""
	}

	return node.GetPath()
}

func (self *FileTreeViewModel) SetTree() {
	newFiles := self.GetAllFiles()
	selectedNode := self.GetSelected()

	// for when you stage the old file of a rename and the new file is in a collapsed dir
	for _, file := range newFiles {
		if selectedNode != nil && selectedNode.path != "" && file.PreviousPath == selectedNode.path {
			self.ExpandToPath(file.Path)
		}
	}

	prevNodes := self.GetAllItems()
	prevSelectedLineIdx := self.GetSelectedLineIdx()

	self.IFileTree.SetTree()
	self.invalidateDisplayCache()

	if selectedNode != nil {
		newNodes := self.GetAllItems()
		newIdx := self.findNewSelectedIdx(prevNodes[prevSelectedLineIdx:], newNodes)
		if newIdx != -1 && newIdx != prevSelectedLineIdx {
			self.SetSelection(newIdx)
		}
	}

	self.ClampSelection()
}

// Let's try to find our file again and move the cursor to that.
// If we can't find our file, it was probably just removed by the user. In that
// case, we go looking for where the next file has been moved to. Given that the
// user could have removed a whole directory, we continue iterating through the old
// nodes until we find one that exists in the new set of nodes, then move the cursor
// to that.
// prevNodes starts from our previously selected node because we don't need to consider anything above that
func (self *FileTreeViewModel) findNewSelectedIdx(prevNodes []*FileNode, currNodes []*FileNode) int {
	getPaths := func(node *FileNode) []string {
		if node == nil {
			return nil
		}
		if node.File != nil && node.File.IsRename() {
			return node.File.Names()
		}
		return []string{node.path}
	}

	getSections := func(node *FileNode) map[FileSectionKind]struct{} {
		sections := map[FileSectionKind]struct{}{}
		if node == nil {
			return sections
		}

		for _, kind := range fileSectionOrder {
			if node.SomeFile(func(file *models.File) bool { return fileSectionForFile(file) == kind }) {
				sections[kind] = struct{}{}
			}
		}

		return sections
	}

	sectionsOverlap := func(a, b map[FileSectionKind]struct{}) bool {
		if len(a) == 0 || len(b) == 0 {
			return true
		}

		for kind := range a {
			if _, ok := b[kind]; ok {
				return true
			}
		}

		return false
	}

	for _, prevNode := range prevNodes {
		selectedPaths := getPaths(prevNode)
		selectedSections := getSections(prevNode)

		for idx, node := range currNodes {
			paths := getPaths(node)
			currSections := getSections(node)
			if !sectionsOverlap(selectedSections, currSections) {
				continue
			}

			// If you started off with a rename selected, and now it's broken in two, we want you to jump to the new file, not the old file.
			// This is because the new should be in the same position as the rename was meaning less cursor jumping
			foundOldFileInRename := prevNode.File != nil && prevNode.File.IsRename() && node.path == prevNode.File.PreviousPath
			foundNode := utils.StringArraysOverlap(paths, selectedPaths) && !foundOldFileInRename
			if foundNode {
				return idx
			}
		}
	}

	return -1
}

func (self *FileTreeViewModel) SetStatusFilter(filter FileTreeDisplayFilter) {
	self.IFileTree.SetStatusFilter(filter)
	self.invalidateDisplayCache()
	self.IListCursor.SetSelection(0)
}

// If we're going from flat to tree we want to select the same file.
// If we're going from tree to flat and we have a file selected we want to select that.
// If instead we've selected a directory we need to select the first file in that directory.
func (self *FileTreeViewModel) ToggleShowTree() {
	selectedNode := self.GetSelected()

	self.IFileTree.ToggleShowTree()
	self.invalidateDisplayCache()

	if selectedNode == nil {
		return
	}
	path := selectedNode.path

	if self.InTreeMode() {
		self.ExpandToPath(path)
	} else if len(selectedNode.Children) > 0 {
		path = selectedNode.GetLeaves()[0].path
	}

	index, found := self.GetIndexForPath(path)
	if found {
		self.SetSelectedLineIdx(index)
	}
}

func (self *FileTreeViewModel) CollapseAll() {
	selectedNode := self.GetSelected()

	self.IFileTree.CollapseAll()
	self.invalidateDisplayCache()
	if selectedNode == nil {
		return
	}

	topLevelPath := strings.Split(selectedNode.path, "/")[0]
	index, found := self.GetIndexForPath(topLevelPath)
	if found {
		self.SetSelectedLineIdx(index)
	}
}

func (self *FileTreeViewModel) ExpandAll() {
	selectedNode := self.GetSelected()

	self.IFileTree.ExpandAll()
	self.invalidateDisplayCache()

	if selectedNode == nil {
		return
	}

	index, found := self.GetIndexForPath(selectedNode.path)
	if found {
		self.SetSelectedLineIdx(index)
	}
}

func (self *FileTreeViewModel) ToggleCollapsed(path string) {
	self.IFileTree.ToggleCollapsed(path)
	self.invalidateDisplayCache()
}

func (self *FileTreeViewModel) GetSections() []FileSection {
	self.ensureDisplayCache()

	self.RLock()
	defer self.RUnlock()

	sections := make([]FileSection, len(self.sectionsCache))
	for i, section := range self.sectionsCache {
		nodes := append([]*FileNode(nil), section.Nodes...)
		sections[i] = FileSection{Kind: section.Kind, Nodes: nodes}
	}

	return sections
}

func (self *FileTreeViewModel) DisplayRoot() *Node[models.File] {
	self.ensureDisplayCache()

	self.RLock()
	defer self.RUnlock()

	if self.displayRootCache == nil {
		return &Node[models.File]{}
	}

	return self.displayRootCache
}

func (self *FileTreeViewModel) GetAllItems() []*FileNode {
	self.ensureDisplayCache()

	self.RLock()
	defer self.RUnlock()

	return append([]*FileNode(nil), self.itemsCache...)
}

func (self *FileTreeViewModel) Get(index int) *FileNode {
	self.ensureDisplayCache()

	self.RLock()
	defer self.RUnlock()

	if index < 0 || index >= len(self.itemsCache) {
		return nil
	}

	return self.itemsCache[index]
}

func (self *FileTreeViewModel) Len() int {
	self.ensureDisplayCache()

	self.RLock()
	defer self.RUnlock()

	return len(self.itemsCache)
}

func (self *FileTreeViewModel) GetIndexForPath(path string) (int, bool) {
	self.ensureDisplayCache()

	self.RLock()
	defer self.RUnlock()

	for idx, node := range self.itemsCache {
		if node != nil && node.GetPath() == path {
			return idx, true
		}
	}

	return 0, false
}

func (self *FileTreeViewModel) invalidateDisplayCache() {
	self.Lock()
	defer self.Unlock()

	self.sectionsCache = nil
	self.itemsCache = nil
	self.displayRootCache = nil
}

func (self *FileTreeViewModel) ensureDisplayCache() {
	self.RLock()
	cached := self.sectionsCache != nil
	self.RUnlock()
	if cached {
		return
	}

	self.Lock()
	defer self.Unlock()

	if self.sectionsCache != nil {
		return
	}

	self.buildDisplayStructure()
}

func (self *FileTreeViewModel) buildDisplayStructure() {
	baseRoot := self.IFileTree.GetRoot()
	if baseRoot == nil {
		self.sectionsCache = []FileSection{}
		self.itemsCache = []*FileNode{}
		self.displayRootCache = &Node[models.File]{}
		return
	}

	rawRoot := baseRoot.Raw()
	if rawRoot == nil {
		self.sectionsCache = []FileSection{}
		self.itemsCache = []*FileNode{}
		self.displayRootCache = &Node[models.File]{}
		return
	}

	collapsedPaths := self.CollapsedPaths()

	displayRoot := &Node[models.File]{}
	sections := []FileSection{}
	items := []*FileNode{}

	for _, kind := range fileSectionOrder {
		sectionRoot := self.cloneSection(rawRoot, kind, collapsedPaths)
		if sectionRoot == nil {
			continue
		}

		flattened := sectionRoot.Flatten(collapsedPaths)
		if len(flattened) <= 1 {
			continue
		}

		displayRoot.Children = append(displayRoot.Children, sectionRoot.Children...)

		fileNodes := lo.Map(flattened[1:], func(node *Node[models.File], _ int) *FileNode {
			return NewFileNode(node)
		})

		sections = append(sections, FileSection{Kind: kind, Nodes: fileNodes})
		items = append(items, fileNodes...)
	}

	self.sectionsCache = sections
	self.itemsCache = items
	self.displayRootCache = displayRoot
}

func (self *FileTreeViewModel) cloneSection(node *Node[models.File], kind FileSectionKind, collapsedPaths *CollapsedPaths) *Node[models.File] {
	if node == nil {
		return nil
	}

	if node.File != nil {
		if fileSectionForFile(node.File) != kind {
			return nil
		}

		return &Node[models.File]{
			File:             node.File,
			path:             node.path,
			CompressionLevel: node.CompressionLevel,
		}
	}

	if !nodeContainsSection(node, kind) {
		return nil
	}

	clone := &Node[models.File]{
		path:             node.path,
		CompressionLevel: node.CompressionLevel,
	}

	if collapsedPaths.IsCollapsed(node.GetInternalPath()) {
		return clone
	}

	for _, child := range node.Children {
		if childClone := self.cloneSection(child, kind, collapsedPaths); childClone != nil {
			clone.Children = append(clone.Children, childClone)
		}
	}

	return clone
}

func nodeContainsSection(node *Node[models.File], kind FileSectionKind) bool {
	return node.SomeFile(func(file *models.File) bool {
		return fileSectionForFile(file) == kind
	})
}

func fileSectionForFile(file *models.File) FileSectionKind {
	switch {
	case file.HasMergeConflicts:
		return FileSectionStaged
	case file.HasStagedChanges:
		return FileSectionStaged
	case !file.Tracked:
		return FileSectionUntracked
	case file.HasUnstagedChanges:
		return FileSectionUnstaged
	default:
		return FileSectionUnstaged
	}
}
