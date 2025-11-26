# treefiddy
WIP but functional tree file explorer. Written so I could use as the file picker in [zide](https://github.com/josephschmitt/zide). Existing options didn't play well with double-clicking via mouse to open the file in the editor.

## Features
  * Tree view rooted to path provided by the first argument (or cwd if none provided)
    * The tree is _never_ replaced by navigation. This is to echo behavior of IDEs such as VSCode.
  * Double-click items to open with $EDITOR
  * Arrow keys/hjkl + enter to open with $EDITOR
  * Symlink following, wow!

This does not use xdg-open or otherwise as of yet, it only calls `$EDITOR <file>`. I'll add such features as they become necessary.
