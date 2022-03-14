# mp3

The purpose of the **mp3** project is to help manage _mp3_ sound files in
Windows. It has three subcommands:

## ls

The **ls** subcommand provides a means for listing mp3 files. It can list
artists, albums, and tracks, governed by these command line arguments:

1. **-artist** List artist names. **True** by default
2. **-album** List album names. **True** by default
3. **-track** List track names. **True** by default

In addition, you can specify that track listings can be in numeric or alphabetic
order:

1. **-sort** Allowed values are **numeric** (default) and **alpha**

If **numeric** sorting is reqested, track and album listing must both be
enabled; otherwise, it makes no sense.

If any value other than **numeric** or **alpha** is used, it will be replaced by
an appropriate value as follows:

1. If tracks are not listed, the value is ignored.
2. If tracks are listed and albums are listed, the value is replaced by
   **numeric**
3. If tracks are listed but albums are not listed, the value is replaced by
   **alpha**

Track and album names can be annotated with the boolean **-annotate** flag,
which is **false** by default. If **true**, the following annotations are
provided:

1. Album names will include the recording artist if artists are not included
   (**-artist=false**)
2. Track names will include the album name if albums are not included
   (**-album=false**), and the recording artist if artists are also not included
   (**-artist=false**)

## check

The **check** subcommand provides a means to run various checks on the mp3 files
and their directories, governed by these command line arguments:

1. **-empty** Check for empty _artist_ and _album_ directories. **False** by
   default. If **true**, ignores the **-albums** and **-artists** filter
   settings. An _artist_ directory is any subdirectory of the **-topDir**
   directory, and is considered empty if it contains no subdirectories. An album
   directory is any subdirectory of an artist directory and is considered empty
   if it contains no mp3 files.
2. **-gaps** Check for gaps in the numbered mp3 files in an _album_ directory.
   **False** by default. The assumption is that the mp3 files in an album
   directory are numbered as tracks, starting with **1** and ending with **N**
   where **N** is the number of mp3 files in the directory. If any mp3 files
   have an associated track number outside the range of **1..N**, those are
   listed in the output, as are those track numbers in the expected range that
   are not associated with any mp3 files in the directory.
3. **-integrity** Check for differences between the _tags_ in an mp3 file and
   the associated file and directory names. **True** by default. Each mp3 file
   is expected to have a file name consisting of a track number, a space, and a
   track title; the file's _TRCK_ (track number/position in set) tag and _TIT2_
   (title/songname/content description) tag are expected to match the file's
   track number and track title. The file's _TALB_ (album/movie/show title) tag
   is expected to match the name of the _album_ directory that contains the
   file. The file's _TPE1_ (lead artist/lead performer/soloist/performing group)
   tag is expected to match the name of the _artist_ directory that contains the
   _album_ directory that contains the file. All differences found are output.

## repair

### Common Arguments

These arguments are common to all subcommands:

1. **-topDir** The directory whose subdirectories are artist names. By default,
   this is **$HOMEPATH\Music**.
2. **-ext** The extension used to identify music files. By default, this is
   **.mp3**
3. **-albums** Filter for which albums to process. By default, **'.*'**
4. **-artists** Filter foe which artists to process. By default, **'.*'**

## Environment

**mp3** depends on the following environment variables being set:

1. **%TMP%** or **%TEMP%** - the system temporary directory. **%TMP%** is
   checked first, and, if not found, then **%TEMP%** is checked.
1. **%HOMEPATH%** - the user's home directory.

## Dependencies

The following third party libraries are used:

* [https://github.com/bogem/id3v2](https://github.com/bogem/id3v2) is used for
  reading and writing the mp3 file metadata.
* [https://github.com/sirupsen/logrus](https://github.com/sirupsen/logrus) is
  used for logging.
* [https://github.com/utahta/go-cronowriter](https://github.com/utahta/go-cronowriter)
  is used for log rotation.

In addition, [https://libs.garden/](https://libs.garden/) is used to find
suitable libraries.
