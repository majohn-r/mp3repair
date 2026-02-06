# Changelog

This project uses [semantic versioning](https://semver.org/); be aware that, until the major version becomes non-zero,
[this proviso](https://semver.org/#spec-item-4) applies.

Key to symbols

- ❗ breaking change
- 🐛 bug fix
- ⚠️ change in behavior, may surprise the user
- 😒 change is invisible to the user
- 🆕 new feature

## v0.45.4

_pre-release `2026-02-03`_

- 🐛 really fix [issue 210](https://github.com/majohn-r/mp3repair/issues/210)!

## v0.45.3

_pre-release `2026-02-03`_

- 🐛 fix [issue 210](https://github.com/majohn-r/mp3repair/issues/210)

## v0.45.2

_pre-release `2026-02-01`_

- 🐛 fixed missing application name in help output.
- 🐛 first shot at a user-focused README.
- 😒 updated dependencies.

## v0.45.1

_pre-release `2026-01-24`_

- 😒 updated dependencies.

## v0.45.0

_pre-release `2025-11-01`_

- ⚠️ fixed [issue 206](https://github.com/majohn-r/mp3repair/issues/206) to use the `XDG` music dir instead of a
command-line `--topDir` argument.

## v0.44.4

_pre-release `2025-10-31`_

- ⚠️ log files used to be written to a temp folder (typically `%HOME%\AppData\Local\Temp\mp3repair\logs`); they're now
written to `%HOME%\AppData\Local\mp3repair\logs`.

## v0.44.3

_pre-release `2025-10-30`_

- ⚠️ the configuration file `defaults.yaml` used to be located in `%HOME%\AppData\Roaming\mp3repair`; it is now located
in `%HOME%\AppData\Local\mp3repair`.

## v0.44.2

_pre-release `2025-10-29`_

- ⚠️ the dirty file that is used to determine whether the Windows Media Player library needs to be reset now uses
[XDG](https://specifications.freedesktop.org/basedir-spec/latest/) based code to locate the file. This addresses
[issue 205](https://github.com/majohn-r/mp3repair/issues/205).

## v0.44.1

_pre-release `2025-10-09`_

- 😒 updated dependencies.

## v0.44.0

_pre-release `2026-08-31`_

- 🐛 added **--maxOpenFiles** argument to the **check** and **repair** commands, which allows the user to specify the
maximum number of simultaneously open files; the provided value must be in the range of 1 to 32767, with the default
value being 1000.
- 🐛 fixed a problem where user values for integer-valued command flags, which are always supposed to be bounded by
minimum and maximum values, were blindly accepted even if they fell outside the prescribed bounds. This affected the
**resetLibrary** command's **timeout** flag and the newly created **--maxOpenFiles** flag. In the case of the
**--maxOpenFiles** flag, a user could specify a very large value, which put the program under a great deal of strain as
it struggled to create a slice of channels that large, or 0, which caused the **check** or **repair** command to freeze
up as code waited for a slot to insert a channel into a 0-sized slice of channels, or a negative value, which caused a
panic as the program attempted to instantiate a slice of channels with a negative size.
- 😒 cleaned up IDE-generated complaints (unsorted imports and a function name that did not need to be capitalized).
- 😒 updated code dependencies.

## v0.43.9

_pre-release `2026-08-29`_

- 😒 updated dependencies, including go 1.25.

## v0.43.6

_pre-release `2026-07-20`_

- 😒 updated dependencies.

## v0.43.5

_pre-release `2025-03-10`_

- 🐛 fixed bug that occurred when the program is opened with elevated privileges, as with the new Windows 11 `sudo`
command or from a window running with elevated privileges, in which, as the program exits, it prompts the user for an
unnecessary 'press enter'. The program now only presents that prompt if it had relaunched itself in a new window with
elevated privileges.

## v0.43.4

_pre-release `2025-02-23`_

- 🐛 fixed bugs handling the processing of non-ASCII characters in file names with respect to the ID3V1 metadata

## v0.43.3

_pre-release `2025-02-20`_

- 🐛 fixed bug where metadata differences were found by **`check -f`** but insufficient detail was given to make it
obvious what to fix (sometimes, for instance, problems exist in the file space, not metadata)

## v0.43.2

_pre-release `2025-02-20`_

- 🐛 fixed bug where repair would fail (be blocked) if a file under repair had been repaired previously and its backup
file had not been cleaned up

## v0.43.0

_pre-release `2024-08-25`_

- ⚠️clean up error output where the user is offered a choice of options to pursue in order to fix their problem

## v0.42.0

_pre-release `2024-08-22`_

- ❗remove support for **resetLibrary** command options **--extension**, **--metadataDir**, and
**--service**

## v0.41.0

_pre-release `2024-08-11`_

- ❗rename **resetDatabase** command to **resetLibrary**

## v0.40.0

_pre-release `2024-08-07`_

- ❗remove support for **list** command **--details** option
- 🐛improve look of output for the **list** command with **--diagnostic** option
- 🐛improve output of **MCDI** field output for the **list** command with **--diagnostic** option
- 🆕add support for pretty output of **APIC** fields for the list command with **--diagnostic** option

## v0.39.1

_pre-release `2024-07-17`_

- 🐛logging bug that injected seemingly random whitespace into the beginning of each record

## v0.39.0

_pre-release `2024-07-16`_

- 🐛minor logging changes, nothing a user is likely to notice

## v0.38.0

_pre-release `2024-07-08`_

- 🆕add **--style** option to **about** command
- 🐛add example to help for the **repair** command's **--dryRun** option

## v0.37.1

_pre-release `2024-06-07`_

- 🐛several **list** command option combinations resulted in failure: **-lrt --byTitle**, **-lrt --byNumber**,
**-lt --byTitle**, **-lt --byNumber**, **-rt --byTitle**, and **-t --byTitle**

## v0.37.0

_pre-release `2024-05-10`_

- 🆕in the **check** command output, when the same concern applies to each track in an album, or each album by an
artist, collapse those identical concerns into a single concern indicating that it applies to each track or to each
album
- 🐛recognize when track metadata is not found in an mp3 file and say so, clearly
- 🐛when the year field in an mp3 file's ID3V2 metadata contains more than just a 4-digit year, e.g.,
_"1969 (2019)"_, do a better job comparing it to the file's ID3V1 year field (which is constrained to be 4 digits)

## v0.36.3

_pre-release `2024-04-05`_

- 😒no significant user-facing changes

## v0.36.2

_pre-release `2024-04-01`_

- 🐛recognize when the user declines to run with elevated privileges
- 🐛when stopping the Windows Media Player sharing service fails, stop advising the user to run with elevated
privileges if they already are doing so

## v0.36.1

_pre-release `2024-03-30`_

- ❗rename program from **mp3** to **mp3repair**

## v0.35.1

_pre-release `2024-03-03`_

- 🆕add details about whether stdin, stderr, and stdout redirection to the **about** command output 

## v0.35.0

_pre-release `2024-03-03`_

- 🆕program runs with elevated permissions when possible
- 🆕add log file location, config file location (and whether it exists), whether the program is running with
elevated permissions (and, if not, why not) to the **about** command output

## v0.34.0

_pre-release `2024-02-16`_

- 🆕add help to all commands and to the program as a whole
- 🆕add icon to the program
- ❗command options now begin with `--` instead of `-`, but abbreviated options now exist as well with
single dash prefix

## v0.33.8

_pre-release `2023-10-04`_

- 😒no significant user-facing changes

## v0.33.7

_pre-release `2023-02-09`_

- 😒no significant user-facing changes

## v0.33.6

_pre-release `2023-02-07`_

- 😒no significant user-facing changes

## v0.33.4 

_pre-release `2022-12-17`_

- 🐛changed the labeling in the progress bar to indicate speed in tracks per second

## v0.33.3

_pre-release `2022-12-08`_

- 🆕first pre-release; reasonably functional