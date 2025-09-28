package file

import (
	"github.com/jesseduffield/lazygit/pkg/config"
	. "github.com/jesseduffield/lazygit/pkg/integration/components"
)

var DiscardStagedChanges = NewIntegrationTest(NewIntegrationTestArgs{
	Description:  "Discarding staged changes",
	ExtraCmdArgs: []string{},
	Skip:         false,
	SetupConfig: func(config *config.AppConfig) {
	},
	SetupRepo: func(shell *Shell) {
		shell.CreateFileAndAdd("fileToRemove", "original content")
		shell.CreateFileAndAdd("file2", "original content")
		shell.Commit("first commit")

		shell.CreateFile("file3", "original content")
		shell.UpdateFile("fileToRemove", "new content")
		shell.UpdateFile("file2", "new content")
	},
	Run: func(t *TestDriver, keys config.KeybindingConfig) {
		t.Views().Files().
			IsFocused().
			Lines(
				Equals("--- Unstaged changes ---"),
				Equals("▼ /").IsSelected(),
				Equals("   M file2"),
				Equals("   M fileToRemove"),
				Equals("--- only untracked ---"),
				Equals("▼ /"),
				Equals("  ?? file3"),
			).
			NavigateToLine(Contains(`fileToRemove`)).
			PressPrimaryAction().
			Lines(
				Equals("--- Staged changes ---"),
				Equals("▼ /").IsSelected(),
				Equals("  M  fileToRemove").IsSelected(),
				Equals("--- Unstaged changes ---"),
				Equals("▼ /"),
				Equals("   M file2"),
				Equals("--- only untracked ---"),
				Equals("▼ /"),
				Equals("  ?? file3"),
			).
			Press(keys.Files.ViewResetOptions)

		t.ExpectPopup().Menu().Title(Equals("")).Select(Contains("Discard staged changes")).Confirm()

		// staged file has been removed
		t.Views().Files().
			Lines(
				Equals("--- Unstaged changes ---"),
				Equals("▼ /"),
				Equals("   M file2"),
				Equals("--- only untracked ---"),
				Equals("▼ /"),
				Equals("  ?? file3").IsSelected(),
			)

		// the file should have the same content that it originally had, given that that was committed already
		t.FileSystem().FileContent("fileToRemove", Equals("original content"))
	},
})
