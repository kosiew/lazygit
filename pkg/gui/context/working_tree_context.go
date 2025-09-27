package context

import (
	"fmt"
	"strings"

	"github.com/jesseduffield/gocui"
	"github.com/jesseduffield/lazygit/pkg/commands/models"
	"github.com/jesseduffield/lazygit/pkg/gui/filetree"
	"github.com/jesseduffield/lazygit/pkg/gui/presentation"
	"github.com/jesseduffield/lazygit/pkg/gui/presentation/icons"
	"github.com/jesseduffield/lazygit/pkg/gui/types"
	"github.com/samber/lo"
)

type WorkingTreeContext struct {
	*filetree.FileTreeViewModel
	*ListContextTrait
	*SearchTrait
}

var (
	_ types.IListContext       = (*WorkingTreeContext)(nil)
	_ types.ISearchableContext = (*WorkingTreeContext)(nil)
)

func NewWorkingTreeContext(c *ContextCommon) *WorkingTreeContext {
	viewModel := filetree.NewFileTreeViewModel(
		func() []*models.File { return c.Model().Files },
		c.Common,
		c.UserConfig().Gui.ShowFileTree,
	)

	getDisplayStrings := func(_ int, _ int) [][]string {
		showFileIcons := icons.IsIconEnabled() && c.UserConfig().Gui.ShowFileIcons
		showNumstat := c.UserConfig().Gui.ShowNumstatInFilesView
		lines := presentation.RenderFileTree(viewModel, c.Model().Submodules, showFileIcons, showNumstat, &c.UserConfig().Gui.CustomIcons, c.UserConfig().Gui.ShowRootItemInFileTree)
		return lo.Map(lines, func(line string, _ int) []string {
			return []string{line}
		})
	}

	getNonModelItems := func() []*NonModelItem {
		sections := viewModel.GetSections()
		if len(sections) == 0 {
			return nil
		}

		labels := map[filetree.FileSectionKind]string{
			filetree.FileSectionStaged:   c.Tr.StagedChanges,
			filetree.FileSectionUnstaged: c.Tr.UnstagedChanges,
		}

		untrackedLabel := c.Tr.FilterUntrackedFiles
		trimmed := strings.Trim(c.Tr.FilterLabelUntrackedFiles, "()（）")
		if value := strings.TrimSpace(trimmed); value != "" {
			untrackedLabel = value
		}
		labels[filetree.FileSectionUntracked] = untrackedLabel

		items := []*NonModelItem{}
		runningIndex := 0
		for _, section := range sections {
			count := len(section.Nodes)
			if count == 0 {
				continue
			}

			label, ok := labels[section.Kind]
			if ok {
				items = append(items, &NonModelItem{
					Index:   runningIndex,
					Content: fmt.Sprintf("--- %s ---", label),
				})
			}

			runningIndex += count
		}

		return items
	}

	ctx := &WorkingTreeContext{
		SearchTrait:       NewSearchTrait(c),
		FileTreeViewModel: viewModel,
		ListContextTrait: &ListContextTrait{
			Context: NewSimpleContext(NewBaseContext(NewBaseContextOpts{
				View:       c.Views().Files,
				WindowName: "files",
				Key:        FILES_CONTEXT_KEY,
				Kind:       types.SIDE_CONTEXT,
				Focusable:  true,
			})),
			ListRenderer: ListRenderer{
				list:              viewModel,
				getDisplayStrings: getDisplayStrings,
				getNonModelItems:  getNonModelItems,
			},
			c: c,
		},
	}

	ctx.GetView().SetOnSelectItem(ctx.SearchTrait.onSelectItemWrapper(ctx.OnSearchSelect))

	return ctx
}

func (self *WorkingTreeContext) ModelSearchResults(searchStr string, caseSensitive bool) []gocui.SearchPosition {
	return nil
}
