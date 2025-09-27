package file

import (
	"fmt"

	"github.com/jesseduffield/lazygit/pkg/config"
	. "github.com/jesseduffield/lazygit/pkg/integration/components"
)

var StageAndUnstagePreservesSelection = NewIntegrationTest(NewIntegrationTestArgs{
	Description:  "staging and immediately unstaging a file keeps it selected when scroll preservation is active",
	ExtraCmdArgs: []string{},
	Skip:         false,
	SetupConfig:  func(config *config.AppConfig) {},
	SetupRepo: func(shell *Shell) {
		shell.CreateFileAndAdd("tracked.txt", "tracked").
			Commit("initial")

		for i := 0; i < 30; i++ {
			shell.CreateFile(fmt.Sprintf("file-%02d.txt", i), "content")
		}
	},
	Run: func(t *TestDriver, keys config.KeybindingConfig) {
		filesView := t.Views().Files().
			Focus().
			SetOriginY(5).
			NavigateToLine(Contains("file-15.txt")).
			SelectedLine(Contains("?? file-15.txt"))

		filesView.
			PressPrimaryAction().
			SelectedLine(Contains("A  file-15.txt"))

		filesView.
			PressPrimaryAction().
			SelectedLine(Contains("?? file-15.txt"))
	},
})
