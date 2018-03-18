# amr2wav
convert wechat AMR audio files to wav format

Decoder code is from OpenCore-AMR from the mirror https://github.com/VFR-maniac/opencore-amr. The files are copied into one directory to make ```go build``` easier.

Please see ```amr2wav.go``` for details.

# Build
```go build```
Tested under MacOS X

# Usage
```Usage: ./amr2wav inputfile outputfile```

# License
Decoder code is under its original license.
amr2wav.go is under the Apache license
