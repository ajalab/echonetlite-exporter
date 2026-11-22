# ECHONET Lite Machine Readable Appendix (MRA)

The ECHONET Consortium publishes the Machine Readable Appendix as a JSON reference for the Detailed Requirements for ECHONET Device Objects.

## Download

Run the following command from the repository root. It requires `curl`, `shasum`, and `unzip`.

```sh
./third_party/echonet-lite-mra/fetch.sh
```

The script downloads the official ZIP, verifies its SHA-256 checksum, and extracts it to `third_party/echonet-lite-mra/MRA_v1.4.0/`. If the data is already present, the script prints its path. Git ignores the extracted files.

## Sources and copyright

The official archive is labeled v1.4.0 and corresponds to Appendix Release R rev.3 (data format v1.2.0). Its `metaData.json` lists `dataVersion` as `1.3.2` and both `note.ja` and `note.en` as `Machine Readable Appendix V1.3.2`. Extraction preserves the original directory structure and JSON contents.

Copyright notice from `metaData.json`

> (C) 2021-2026 Kanagawa Institute of Technology, ECHONET Consortium

The repository's MIT license covers the project's code and documentation. [MIT license](../../LICENSE)

- [Official release page](https://echonet.jp/spec_mra_rr3/)
- [Official archive (MRA_v1.4.0.zip)](https://echonet.jp/wp/wp-content/uploads/pdf/General/Standard/MRA/MRA_v1.4.0.zip)
- [Release notes v1.4.0 (Japanese)](https://echonet.jp/wp/wp-content/uploads/pdf/General/Standard/MRA/MRA_releasenote_jp_v1.4.0.pdf)
- [Guidebook v1.2.0 (Japanese)](https://echonet.jp/wp/wp-content/uploads/pdf/General/Standard/MRA/MRA_guidebook_jp_v1.2.0.pdf)
