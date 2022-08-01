# mp3

- [mp3](#mp3)
  - [Purpose](#purpose)
  - [Commands](#commands)
    - [check](#check)
    - [ls](#ls)
    - [postRepair](#postrepair)
    - [repair](#repair)
    - [resetDatabase](#resetdatabase)
  - [Command Arguments](#command-arguments)
    - [Common Command Arguments](#common-command-arguments)
    - [Specifying Command Line Arguments](#specifying-command-line-arguments)
    - [Overriding Default Arguments](#overriding-default-arguments)
    - [Argument Values](#argument-values)
      - [Using Environment Variables](#using-environment-variables)
      - [File Separators](#file-separators)
      - [Numeric Values](#numeric-values)
      - [Boolean Values](#boolean-values)
  - [Environment](#environment)
  - [Dependencies](#dependencies)

## Purpose

The purpose of the **mp3** project is to help manage _mp3_ sound files in
Windows. It supports five commands:

## Commands

The **mp3** program supports five commands.

### check

The **check** command provides a means to run various checks on the mp3 files
and their directories, governed by these command arguments:

1. **-empty** Check for empty _artist_ and _album_ directories. **False** by
   default. If **true**, **mp3** ignores the **-albumFilter** and
   **-artistFilter** settings. An _artist_ directory is any subdirectory of the
   **-topDir** directory, and **mp3** considers it to be empty if it contains no
   subdirectories. An album directory is any subdirectory of an artist directory
   and **mp3** considers it empty if it contains no mp3 files.
2. **-gaps** Check for gaps in the numbered mp3 files in an _album_ directory.
   **False** by default. **mp3** assumes that the mp3 files in an album
   directory are numbered as tracks, starting with **1** and ending with **N**
   where **N** is the number of mp3 files in the directory. If any mp3 files
   have an associated track number outside the range of **1..N**, **mp3** lists
   them in the output, as well as any track numbers in the expected range that
   are not associated with any mp3 files in the directory.
3. **-integrity** Check for differences between the _tags_ in an mp3 file and
   the associated file and directory names. **True** by default. **mp3** expects
   each mp3 file to have a file name consisting of a track number, a delimiter
   character (space or punctuation mark), and a track title. **mp3** expects the
   file's _TRCK_ (track number/position in set) tag and _TIT2_
   (title/songname/content description) tag to match the file's track number and
   track title. **mp3** expects the file's _TALB_ (album/movie/show title) tag
   to match the name of the _album_ directory that contains the file. **mp3**
   expects the file's _TPE1_ (lead artist/lead performer/soloist/performing
   group) tag to match the name of the _artist_ directory that contains the
   _album_ directory that contains the file. The matching takes into account the
   fact that the tags may contain characters that are illegal in file and
   directory names and are, therefore, replaced by other characters (typically
   punctuation marks). All differences found are output.

### ls

The **ls** command provides a means for listing mp3 files. It can list artists,
albums, and tracks, governed by these command arguments:

1. **-includeArtists** List artist names. **True** by default
2. **-includeAlbums** List album names. **True** by default
3. **-includeTracks** List track names. **False** by default

In addition, you can specify that track listings can be in numeric or alphabetic
order:

1. **-sort** Allowed values are **numeric** (default) and **alpha**

If **numeric** sorting is reqested, track and album listing must both be
enabled; otherwise, it makes no sense.

If any value other than **numeric** or **alpha** is used, **mp3** will be
replace it with an appropriate value as follows:

1. If tracks are not listed, **mp3** ignores the value.
2. If tracks and albums are listed, **mp3** replaces the value with **numeric**.
3. If tracks are listed but albums are not listed, **mp3** replaces the value
   with **alpha**.

Track and album names can be annotated with the boolean **-annotate** flag,
which is **false** by default. If **true**, **mp3** provides the following
annotations:

1. Album names will include the recording artist if artists are not included
   (**-includeArtists=false**)
2. Track names will include the album name if albums are not included
   (**-includeAlbums=false**), and the recording artist if artists are also not included
   (**-includeArtists=false**)

The boolean **-diagnostic** flag adds detailed internal data to each track, if
tracks are listed (**-includeTracks=true**).

### postRepair

The **postRepair** command provides a means to quickly delete the backup
directories created by the **repair** command. It has no command arguments.

### repair

The **repair** command provides a means to repair tracks whose **MP3** _tags_ do
not match the track name, album name, or artist name. It has a single command
argument:

1. **-dryRun** If true, outputs what the **repair** command would fix. **False**
   by default.

### resetDatabase

The **resetDatabase** command provides a means to reset the database that the
Windows Media Player uses to catalogue the albums, artists, and tracks. The
Windows Media Player will not recognize the effects of the **repair** command
until that database is reset. The command has the following command arguments:

1. **-extension** the extension of the files to delete; defaults to **.wmdb**.
2. **-metadata** the directory where the metadata files are found; defaults to
   **%Userprofile%\AppData\Local\Microsoft\Media Player\**.
3. **-service** the name of the media player sharing service, which, if running,
   needs to be stopped before deleting the metadata files; defaults to
   **WMPNetworkSVC**.
4. **-timeout** the time, in seconds, in which the command will attempt to stop
   the media player sharing service before giving up; the minimum value is
   **1**, the maximum value is **60**, and the default value is **10**.

## Command Arguments

The commands take a variety of _command arguments_. These arguments may be
string-valued, numeric-valued, or boolean-valued (true/false). Many of the
command arguments are command-specific and are described with the commands
above.

### Common Command Arguments

These command arguments are common to all commands except the **resetDatabase**
command:

1. **-topDir** The directory whose subdirectories are artist names. By default,
   this is **%HOMEPATH%\Music**.
2. **-ext** The extension used to identify music files. By default, this is
   **.mp3**
3. **-albumFilter** Filter for which albums to process. By default, **'.*'**
4. **-artistFilter** Filter foe which artists to process. By default, **'.*'**

### Specifying Command Line Arguments

Command arguments can be specified on the command line. On the command line,
string-valued and numeric-valued arguments can be entered in any of the
following forms:

- **-argumentName** **argumentValue**
- **-argumentName=argumentValue**
- **--argumentName** **argumentValue**
- **--argumentName=argumentValue**

Boolean (true/false) arguments can be entered in either of the follwing forms if
the argument value is false:

- **-argumentName=false**
- **--argumentName=false**

They can be entered in any of the following forms if the argument value is true:

- **-argumentName**
- **-argumentName=true**
- **--argumentName**
- **--argumentName=true**

### Overriding Default Arguments

The various command arguments have built-in default values; the command
arguments do not need to be specified on the command line unless the user wants
to specify different values for those command arguments.

The user may find that she is constantly overriding some command arguments on
the command line with the same values. The user can simplify their usage and
override the default command argument values by placing a text file named
**defaults.yaml** in the **%APPDATA%\mp3** directory. By default, there is no
such directory, and the user must create it.

The **YAML** format documentation is here: [https://yaml.org/](https://yaml.org/).

The **defaults.yaml** file may contain six blocks, all of which are optional:

1. **check** The **check** block may have up to three boolean key-value pairs,
   with each key controlling the default setting for its corresponding **check**
   command argument:
   1. **empty**
   2. **gaps**
   3. **integrity**
2. **command** The **command** block may have one string key-value pair:
   1. **default** the value of this entry must be one of **check**, **ls**,
      **postRepair**, **repair**, or **resetDatabase**. It causes that command
      to become the default command when no command is specified on the command
      line.
3. **common** The **common** block may have up to four string key-value pairs,
   with each key controlling the default setting for its corresponding
   **common** argument:
   1. **albumFilter**
   2. **artistFilter**
   3. **ext**
   4. **topDir**
4. **ls** The **ls** block may have up to four boolean key-value pairs and one
   string key-value pair, with each key controlling the default setting for its
   corresponding **ls** command argument:
   1. **annotate**
   2. **diagnostic**
   3. **includeAlbums**
   4. **includeArtists**
   5. **includeTracks**
   6. **sort** must be set to **alpha** or **numeric**
5. **repair** The **repair** block may have one boolean key-value pair,
   controlling the default setting for its corresponding **repair** command
   argument:
   1. **dryRun**
6. **resetDatabase** The **resetDatabase** block may have three string key-value
   pairs and on numeric key-value pair, with each key controlling the default
   setting for its corresponding **resetDatabase** command argument:
   1. **extension**
   2. **metadata**
   3. **service**
   4. **timeout**

Here is the **yaml** content corresponding to the standard out of the box
default values:

```yaml
---
check:
    empty:     false
    gaps:      false
    integrity: true
command:
    default: ls
common:
    albumFilter:  .*
    artistFilter: .* 
    ext:          .mp3
    topDir:       $HOMEPATH/Music
ls:
    annotate:       false
    diagnostic:     false
    includeAlbums:  true
    includeArtists: true
    includeTracks:  false
    sort:           numeric
repair:
    dryRun: false
resetDatabase:
    extension: .wmbd
    metadata:  $Userprofile/AppData/Local/Microsoft/Media Player/
    service:   WMPNetworkSVC
    timeout:   10
```

### Argument Values

A few comments concerning argument values:

#### Using Environment Variables

Argument values may contain references to environment variables. These can be
specified either in **Windows** format (**%VAR_NAME%**) or in **\*nix** format
(**\$VAR_NAME**), such as **$APPDATA/mp3** or **%APPDATA%\mp3**.

#### File Separators

File separators, as in a path to the music files, may be forward slashes (**/**) or
backward slashes (**\\**).

#### Numeric Values

Numeric argument values may be specified in decimal (e.g., _1234_), octal (e.g.,
_0622_), or hexadecimal (e.g., _0x2B_), and may be negative.

#### Boolean Values

Boolean argument values are true or false. True values may be specified as _1_,
_t_, _T_, _true_, _TRUE_, or _True_. False values may be specified as _0_, _f_,
_F_, _false_, _FALSE_, or _False_.

## Environment

**mp3** depends on the following environment variables being set:

1. **TMP** or **TEMP** - the system temporary directory. mp3 looks for **TMP**
   first, and, if that variable is not defined, then mp3 looks for **TEMP**. One
   of them must be set so that log files can be written.

## Dependencies

**mp3** uses the following third party libraries:

- Log rotation:
  [https://github.com/utahta/go-cronowriter](https://github.com/utahta/go-cronowriter).
- Logging:
  [https://github.com/sirupsen/logrus](https://github.com/sirupsen/logrus).
- Reading and writing the mp3 file metadata:
  [https://github.com/bogem/id3v2](https://github.com/bogem/id3v2).
- Reading configuration files:
  [https://pkg.go.dev/gopkg.in/yaml.v3](https://pkg.go.dev/gopkg.in/yaml.v3)

In addition, I use [https://libs.garden/](https://libs.garden/) to look for
libraries.
