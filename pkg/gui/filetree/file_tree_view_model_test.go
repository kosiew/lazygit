package filetree

import (
	"testing"

	"github.com/jesseduffield/lazygit/pkg/commands/models"
	"github.com/jesseduffield/lazygit/pkg/common"
	"github.com/stretchr/testify/assert"
)

func TestFileTreeViewModelSectionsTreeMode(t *testing.T) {
	cmn := common.NewDummyCommon()
	cmn.UserConfig().Gui.ShowRootItemInFileTree = false

	files := []*models.File{
		{Path: "dir/staged.txt", HasStagedChanges: true},
		{Path: "dir/unstaged.txt", HasUnstagedChanges: true, Tracked: true},
		{Path: "dir/sub/untracked.txt", HasUnstagedChanges: true, Tracked: false},
	}

	viewModel := NewFileTreeViewModel(func() []*models.File { return files }, cmn, true)
	viewModel.SetTree()

	sections := viewModel.GetSections()
	if assert.Len(t, sections, 3) {
		assert.Equal(t, FileSectionStaged, sections[0].Kind)
		assert.Equal(t, []string{"dir", "dir/staged.txt"}, pathsFromNodes(sections[0].Nodes))

		assert.Equal(t, FileSectionUnstaged, sections[1].Kind)
		assert.Equal(t, []string{"dir", "dir/unstaged.txt"}, pathsFromNodes(sections[1].Nodes))

		assert.Equal(t, FileSectionUntracked, sections[2].Kind)
		assert.Equal(t, []string{"dir", "dir/sub", "dir/sub/untracked.txt"}, pathsFromNodes(sections[2].Nodes))
	}

	allItems := viewModel.GetAllItems()
	expected := []string{
		"dir", "dir/staged.txt", // staged section
		"dir", "dir/unstaged.txt", // unstaged section
		"dir", "dir/sub", "dir/sub/untracked.txt", // untracked section
	}
	assert.Equal(t, expected, pathsFromNodes(allItems))
	assert.Equal(t, len(expected), viewModel.Len())
}

func TestFileTreeViewModelSectionsFlatMode(t *testing.T) {
	cmn := common.NewDummyCommon()
	files := []*models.File{
		{Path: "staged.txt", HasStagedChanges: true},
		{Path: "unstaged.txt", HasUnstagedChanges: true, Tracked: true},
		{Path: "untracked.txt", HasUnstagedChanges: true, Tracked: false},
	}

	viewModel := NewFileTreeViewModel(func() []*models.File { return files }, cmn, false)
	viewModel.SetTree()

	sections := viewModel.GetSections()
	if assert.Len(t, sections, 3) {
		assert.Equal(t, FileSectionStaged, sections[0].Kind)
		assert.Equal(t, []string{"staged.txt"}, pathsFromNodes(sections[0].Nodes))

		assert.Equal(t, FileSectionUnstaged, sections[1].Kind)
		assert.Equal(t, []string{"unstaged.txt"}, pathsFromNodes(sections[1].Nodes))

		assert.Equal(t, FileSectionUntracked, sections[2].Kind)
		assert.Equal(t, []string{"untracked.txt"}, pathsFromNodes(sections[2].Nodes))
	}

	assert.Equal(t, []string{"staged.txt", "unstaged.txt", "untracked.txt"}, pathsFromNodes(viewModel.GetAllItems()))
	assert.Equal(t, 3, viewModel.Len())
}

func TestSelectionStaysInSameSectionAfterRefresh(t *testing.T) {
	cmn := common.NewDummyCommon()
	cmn.UserConfig().Gui.ShowRootItemInFileTree = false

	files := []*models.File{
		{Path: "dir/staged.txt", HasStagedChanges: true, Tracked: true},
		{Path: "dir/unstaged.txt", HasUnstagedChanges: true, Tracked: true},
	}

	viewModel := NewFileTreeViewModel(func() []*models.File { return files }, cmn, true)
	viewModel.SetTree()

	allItems := viewModel.GetAllItems()
	unstagedDirIdx := -1
	for i, node := range allItems {
		if node.GetPath() != "dir" {
			continue
		}

		if node.SomeFile(func(file *models.File) bool { return fileSectionForFile(file) == FileSectionUnstaged }) {
			unstagedDirIdx = i
			break
		}
	}

	if unstagedDirIdx == -1 {
		t.Fatalf("failed to find unstaged directory entry")
	}

	viewModel.SetSelectedLineIdx(unstagedDirIdx)

	viewModel.SetTree()

	selected := viewModel.GetSelected()
	if assert.NotNil(t, selected) {
		assert.Equal(t, "dir", selected.GetPath())
		assert.True(t, selected.SomeFile(func(file *models.File) bool { return fileSectionForFile(file) == FileSectionUnstaged }))
		assert.False(t, selected.SomeFile(func(file *models.File) bool { return fileSectionForFile(file) == FileSectionStaged }))
	}
}

func pathsFromNodes(nodes []*FileNode) []string {
	result := make([]string, len(nodes))
	for i, node := range nodes {
		if node == nil {
			result[i] = ""
			continue
		}
		result[i] = node.GetPath()
	}
	return result
}

func TestSelectionStaysOnFileWhenMovingBetweenSections(t *testing.T) {
	cmn := common.NewDummyCommon()

	files := []*models.File{
		{Path: "file.txt", HasUnstagedChanges: true, Tracked: true},
	}

	viewModel := NewFileTreeViewModel(func() []*models.File { return files }, cmn, false)
	viewModel.SetTree()

	idx, found := viewModel.GetIndexForPath("file.txt")
	if !found {
		t.Fatalf("failed to find file in initial tree")
	}

	viewModel.SetSelectedLineIdx(idx)

	if selected := viewModel.GetSelected(); assert.NotNil(t, selected) {
		assert.Equal(t, "file.txt", selected.GetPath())
	}

	files[0].HasUnstagedChanges = false
	files[0].HasStagedChanges = true

	viewModel.SetTree()

	if selected := viewModel.GetSelected(); assert.NotNil(t, selected) {
		assert.Equal(t, "file.txt", selected.GetPath())
		assert.True(t, selected.SomeFile(func(file *models.File) bool { return fileSectionForFile(file) == FileSectionStaged }))
	}

	files[0].HasStagedChanges = false
	files[0].HasUnstagedChanges = true

	viewModel.SetTree()

	if selected := viewModel.GetSelected(); assert.NotNil(t, selected) {
		assert.Equal(t, "file.txt", selected.GetPath())
		assert.True(t, selected.SomeFile(func(file *models.File) bool { return fileSectionForFile(file) == FileSectionUnstaged }))
	}
}
