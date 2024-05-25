Introduction
---
**gopusenc** provides bindings to [libopusenc](https://github.com/xiph/libopusenc/). Number of API calls covered is far 
from complete and the library wasn't tested thoroughly, so use it at your own risk.
Feel free to fork.

Example
---
```go
const SamplesPerChannel = 256

oggFile, _ := os.Create("/home/nvevg/stream.ogg")
encoder := gopusenc.NewEncoder(HzRate, NumberOfChannels, gopusenc.MonoStereo, oggFile)
defer encoder.Finish()

if e := encoder.Init(); e != nil {
	fmt.Fprintf(os.Stderr, "cannot initialize encoder: %s\n", e.Error())
	os.Exit(1)
}

// feed int16 PCM samples
pcms := getI16Pcms(SamplesPerChannel)
if e := encoder.Encode(SamplesPerChannel, pcms); e != nil {
    fmt.Fprintf(os.Stderr, "encoding error: %s\n", e.Error())
}
```