package file

import (
	"fmt"

	"github.com/jesseduffield/lazygit/pkg/config"
	. "github.com/jesseduffield/lazygit/pkg/integration/components"
)

var BackgroundRefreshClampsOrigin = NewIntegrationTest(NewIntegrationTestArgs{
	Description:  "background auto-refresh clamps the files view origin when the list shrinks",
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

		for i := 0; i < 60; i++ {
			shell.CreateFile(fmt.Sprintf("untracked-%02d.txt", i), "content")
		}
	},
	Run: func(t *TestDriver, keys config.KeybindingConfig) {
		filesView := t.Views().Files().Focus()

		filesView.NavigateToLine(Contains("untracked-55.txt"))

		savedOrigin := 0
		filesView.Tap(func() {
			savedOrigin = filesView.OriginY()
			if savedOrigin <= 5 {
				t.Fail(fmt.Sprintf("expected initial origin to be greater than 5 but got %d", savedOrigin))
			}
		})

		for i := 15; i < 60; i++ {
			t.Shell().DeleteFile(fmt.Sprintf("untracked-%02d.txt", i))
		}

		t.Wait(1500)

		filesView.
			ContainsLines(Contains("untracked-14.txt").IsSelected()).
			Tap(func() {
				origin := filesView.OriginY()
				if origin >= savedOrigin {
					t.Fail(fmt.Sprintf("expected files view origin to decrease from %d but got %d", savedOrigin, origin))
				}
				if origin > 5 {
					t.Fail(fmt.Sprintf("expected files view origin to be clamped to <= 5 but got %d", origin))
				}
			})
	},
})
