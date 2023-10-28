# Maintenance of Windows Executable Icons

This document describes the creation and maintenance of icons for embedding into Windows execuatables.

## Creation

Creation of an image file is a matter of choosing a drawing program and generating a file from it. The image should be sized for the desired icon's dimensions (a square, such as 256x256). For the **mp3** icon, I used Windows Paint. I also chose to save the image as a .png (Portable Network Graphics) file.

## Conversion to .ico Format

I use [ImageMagick](https://imagemagick.org/), using this command line:

```bash
convert mp3.png -colors 256 mp3.ico
```

Note: the command line tools are not provided by the default **ImageMagick** installer, but is an  option you need to select when installing **ImageMagick**.

## Creation of **.syso** Files

I used [rsrc](https://github.com/akavel/rsrc), installing it with this command line:

```bash
go install github.com/akavel/rsrc@latest
```

And I ran these commands to produce the **.syso** files:

```bash
rsrc -arch 386 -ico mp3.ico
rsrc -arch amd64 -ico mp3.ico
```

## Embedding in the Windows Executable

Embedding the icons is a simple matter of making sure the **.syso** files are in the same folder as the file that contains the **main()** function, which in this case is the **cmd/mp3** folder. Running the build command:

```bash
./build.sh build
```

automatically incorporates the **.syso** files into the executable file.

## Credit Where Credit is Due

I learned all this [here](https://hjr265.me/blog/adding-icons-for-go-built-windows-executable/).
