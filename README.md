# treefiddy
![tree fiddy](treefiddy.png)

WIP but functional tree file explorer. Written so I could use as the file picker in [zide](https://github.com/josephschmitt/zide). Existing options didn't play well with double-clicking via mouse to open the file with $EDITOR (they all wanted to use xdg-open...).

## Features
  * Tree view rooted to path provided by the first argument (or cwd if none provided)
    * The tree is _never_ replaced by navigation. This is to echo behavior of IDEs such as VSCode.
  * Double-click items to `edit` (open with $EDITOR)
  * Arrow keys/hjkl + enter to `edit` (open with $EDITOR)
  * Symlink following, wow!
  * Edicts
    * Otherwise could be called "commands". These are functions that tie a keyword to an action, such as "edit" opening the selected/clicked file in your $EDITOR, or "open" opening using xdg-open.
  * Actions
    * Change what edict a mouse click does or the enter key over a file.
  * Binds
    * Add/change hotkeys to activate an edict.
  * Shortcuts
    * Add short hand phrasing for edicts, such as "ed" referring to "edit", as useable in the command field.

## Configuration
Upon starting, `~/.config/treefiddy/config.yaml` (or whatever your XDG config dir is) is created with the defaults. Configurable options are actions, binds, shortcuts, and a `use_mouse` field.

