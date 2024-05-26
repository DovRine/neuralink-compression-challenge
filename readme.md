# Neuralink Compression Challenge

content.neuralink.com/compression-challenge/data.zip is one hour of raw electrode recordings from a Neuralink implant.

This Neuralink is implanted in the motor cortex of a non-human primate, and recordings were made while playing a video game, like this.

Compression is essential: N1 implant generates ~200Mbps of eletrode data (1024 electrodes @ 20kHz, 10b resolution) and can transmit ~1Mbps wirelessly.
So > 200x compression is needed.
Compression must run in real time (< 1ms) at low power (< 10mW, including radio).

Neuralink is looking for new approaches to this compression problem, and exceptional engineers to work on it.
If you have a solution, email compression@neuralink.com
Leaderboard
Name 	Compression ratio 	Compressed size 	./encode size 	./decode size
zip 	2.2 	63M 	231K 	480K
Task

Build executables ./encode and ./decode which pass eval.sh. This verifies compression is lossless and measures compression ratio.

Your submission will be scored on the compression ratio it achieves on a different set of electrode recordings.
Bonus points for optimizing latency and power efficiency

Submit with source code and build script. Should at least build on Linux.
Data
$ ls -lah data/
total 143M
193K 0052503c-2849-4f41-ab51-db382103690c.wav
193K 006c6dd6-d91e-419c-9836-c3f320da4f25.wav
...

    Uncompressed monochannel WAV files.
    5 seconds per file.


## NOTE: `eval.sh` is modified from the original to include the bin folder when calling the executables
## tested on arch linux

## requirements

- golang 1.21+

## setup
```bash
chmod +x eval.sh
chmod +x setup.sh
./setup.sh
```

## run
```bash
./eval.sh
```


