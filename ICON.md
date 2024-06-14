# Maintenance of Windows Executable Icons

This document describes the creation and maintenance of icons for embedding into
Windows executables.

## Creation

Creation of an image file is a matter of choosing a drawing program and
generating a file from it. The image should be sized for the desired icon's
dimensions (a square, such as 256x256). For the **mp3repair** icon, I used
Windows Paint. I also chose to save the image as a .png (Portable Network
Graphics) file.

## Conversion to .ico Format

I use [ImageMagick](https://imagemagick.org/), using this command line:

```bash
convert mp3repair.png -colors 256 mp3repair.ico
```

Note: the command line tools are not provided by the default **ImageMagick**
installer, but is an option you need to select when installing **ImageMagick**.

## Embedding in the Windows Executable

Embedding the icon is a simple matter of executing the build command.

```bash
./build.sh build
```

## Credit Where Credit is Due

I learned about embedding icons
[here](https://hjr265.me/blog/adding-icons-for-go-built-windows-executable/). I
subsequently investigated populating other executable application file
properties, and automated the generation and inclusion of the icon and various
additional properties into the executable image, using the
[goversioninfo](https://github.com/josephspurrier/goversioninfo) package.
