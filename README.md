# treefiddy
![tree fiddy](treefiddy.png)

**treefiddy** is a tree-centric and directory-rooted file explorer in the vein of an IDE's file browser/picker. Written so I could use as the file picker in [zide](https://github.com/josephschmitt/zide). Existing options didn't play well with mouse use and/or were only pseudo tree-based at best.

⚠  treefiddy is a work-in-progress and may have breaking changes at any point until official release.

## Features
  * Tree view rooted to path provided by the first argument or cwd.
    * Location can be updated via location navbar.
  * Double-click items to `edit` (open with $EDITOR)
  * Arrow keys/hjkl + enter to `edit` (open with $EDITOR)
  * Symlink following, wow!
  * Edicts -- commands via the edict/status bar at the bottom or activated by actions, binds, or plugins. Generally location aware (e.g., `edit foo.go` whilst in `my/dir` will open `my/dir/foo.go`)
    * edit
      * Open files in $EDITOR
    * open
      * Open files in xdg-open or open
    * create
      * Create a file
    * rename
      * Rename a file
    * remove
      * Remove a file
  * Actions
    * Change what edict a mouse click does or the enter key over a file.
  * Binds
    * Add/change hotkeys to activate an edict.
  * Shortcuts
    * Add short hand phrasing for edicts, such as "ed" referring to "edit", as useable in the command field.
  * Plugins
    * JavaScript-based plugins systems for providing features such as sorting, nerd fonts, or git file status.
    * Flexible system for writing plugin systems for other languages.

## Configuration
Upon starting, `~/.config/treefiddy/config.yaml` (or whatever your XDG config dir is) is created with the defaults. See `config.go` for current options.

