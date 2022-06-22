# mp3

The purpose of the **mp3** project is to help manage _mp3_ sound files in
Windows. It has three subcommands:

## ls

The **ls** subcommand provides a means for listing mp3 files. It can list
artists, albums, and tracks, governed by these command line arguments:

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

## check

The **check** subcommand provides a means to run various checks on the mp3 files
and their directories, governed by these command line arguments:

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

## repair

The **repair** subcommand provides a means to repair tracks whose **MP3** _tags_
do not match the track name, album name, or artist name. It has a single command
line argument:

1. **-dryRun** If true, outputs what the **repair** subcommand would fix.
   **False** by default.

## postRepair

The **postRepair** subcommand provides a means to quickly delete the backup
directories created by the **repair** subcommand. It has no command line
arguments.

### Common Arguments

These arguments are common to all subcommands:

1. **-topDir** The directory whose subdirectories are artist names. By default,
   this is **%HOMEPATH%\Music**.
2. **-ext** The extension used to identify music files. By default, this is
   **.mp3**
3. **-albumFilter** Filter for which albums to process. By default, **'.*'**
4. **-artistFilter** Filter foe which artists to process. By default, **'.*'**

## Overriding Default Arguments

The user can override default arguments by placing a text file named
**defaults.yaml** in the **%APPDATA%\mp3** directory. By default, there is no
such directory, and the user must create it.

The **YAML** format documentation is here: [https://yaml.org/](https://yaml.org/).

The **defaults.yaml** file may contain four blocks:

1. **check** The **check** block may have up to three boolean key-value pairs,
   with each key controlling the default setting for its corresponding **check**
   subcommand argument:
   1. **empty**
   1. **gaps**
   1. **integrity**
1. **common** The **common** block may have up to four string key-value pairs,
   with each key controlling the default setting for its corresponding
   **common** argument:
   1. **albumFilter**
   1. **artistFilter**
   1. **ext**
   1. **topDir**
1. **ls** The **ls** block may have up to four boolean key-value pairs and one
   string key-value pair, with each key controlling the default setting for its
   corresponding **ls** command argument:
   1. **annotate**
   1. **includeAlbums**
   1. **includeArtists**
   1. **includeTracks**
   1. **sort** must be set to **alpha** or **numeric**
1. **repair** The **repair** block may have one boolean key-value pair,
   controlling the default setting for its corresponding **repair** command
   argument:
   1. **dryRun**

Here is the **yaml** content corresponding to the standard out of the box
default values:

```yaml
---
check:
    empty:     false
    gaps:      false
    integrity: true
common:
    albumFilter:  .*
    artistFilter: .* 
    ext:          .mp3
    topDir:       $HOMEPATH/Music
ls:
    annotate:       false
    includeAlbums:  true
    includeArtists: true
    includeTracks:  false
    sort:           numeric
repair:
    dryRun: false
```

## Argument Values

Argument values, such as the **common** **topDir** value, may contain references
to environment variables. These can be specified either in **Windows** format
(**%VAR_NAME%**) or in **\*nix** format (**$VAR_NAME**).

This applies to command line arguments and to defaults defined in
**defaults.yaml**.

## Environment

**mp3** depends on the following environment variables being set:

1. **%TMP%** or **%TEMP%** - the system temporary directory. mp3 looks for
   **%TMP%** first, and, if that variable is not defined, then mp3 looks for
   **%TEMP%**. One of them must be set so that log files can be written.

## Dependencies

**mp3** uses the following third party libraries:

* Log rotation:
  [https://github.com/utahta/go-cronowriter](https://github.com/utahta/go-cronowriter).
* Logging:
  [https://github.com/sirupsen/logrus](https://github.com/sirupsen/logrus).
* Reading and writing the mp3 file metadata:
  [https://github.com/bogem/id3v2](https://github.com/bogem/id3v2).
* Reading configuration files:
  [https://pkg.go.dev/gopkg.in/yaml.v3](https://pkg.go.dev/gopkg.in/yaml.v3)

In addition, I use [https://libs.garden/](https://libs.garden/) to look for
libraries.
