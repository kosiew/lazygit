package file

import (
	"fmt"

	"github.com/jesseduffield/lazygit/pkg/config"
	. "github.com/jesseduffield/lazygit/pkg/integration/components"
)

var BackgroundRefreshPreservesScrollOffset = NewIntegrationTest(NewIntegrationTestArgs{
	Description:  "background auto-refresh updates files while preserving the user's scroll position",
	ExtraCmdArgs: []string{},
	Skip:         false,
	SetupConfig: func(cfg *config.AppConfig) {
		userConfig := cfg.GetUserConfig()
		userConfig.Git.AutoRefresh = true
		userConfig.Refresher.RefreshInterval = 1
	},
	SetupRepo: func(shell *Shell) {
		shell.CreateFile("tracked.txt", "tracked").
			GitAddAll().
			Commit("initial")

		for i := 0; i < 20; i++ {
			shell.CreateFile(fmt.Sprintf("untracked-%02d.txt", i), "content")
		}
	},
	Run: func(t *TestDriver, keys config.KeybindingConfig) {
		filesView := t.Views().Files().Focus()

		savedOrigin := 5
		filesView.SetOriginY(savedOrigin).
			Tap(func() {
				savedOrigin = filesView.OriginY()
			})

		t.Shell().CreateFile("untracked-new.txt", "new content")

		t.Wait(1500)

		filesView.ContainsLines(Contains("untracked-new.txt")).
			Tap(func() {
				origin := filesView.OriginY()
				if origin != savedOrigin {
					t.Fail(fmt.Sprintf("expected files view origin %d but got %d", savedOrigin, origin))
				}
			})

		filesView.SelectedLineIdx(0)
	},
})
