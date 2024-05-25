storj-id is a generic encoder/decoder of binary data, used in storj ecosystem.

Usage: 

```
storj-id 12HShX9CtL9Lm2HwpvyTE1De6ievDf2TVkMgHRGnHsvEQ4pDXMF
Using 12HShX9CtL9Lm2HwpvyTE1De6ievDf2TVkMgHRGnHsvEQ4pDXMF as base58
   base32 VEEDR45NTUKII42ZFTJCOVFFEGPRP5HVA2LJGLZ5HVFPAAAAAAAA
   base58 CNqB2geJc7RKdDBAG63t6oz3f9LPMqhgwJvL2jkTiV1h
   base64 qQg4862dFIRzWSzSJ1SlIZ8X9PUGlpMvPT1K8AAAAAA=
   binary 8��sY,�'T�!�����/==J�
   hex a90838f3ad9d148473592cd22754a5219f17f4f50696932f3d3d4af000000000
   nodeid 12HShX9CtL9Lm2HwpvyTE1De6ievDf2TVkMgHRGnHsvEQ4pDXMF
   path veedr45ntukii42zftjcovffegprp5hva2ljglz5hvfpaaaaaaaa
```

Program will try to parse the input string with multiple algorithm, and if it worked, it re-encode 
with all known decoding type.

