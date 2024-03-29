# w2d
Cli-Tool to convert wikipedia articles to markdown and translates them using DeepL. 

## Why?
- Read an article in multiple languages to get new perspectives on a topic
- Send translated wikipedia articles to friends where a topic is not available in their language
- Read wikipedia articles in your terminal

## Requirements
To use the translation functionality you need to [register](https://www.deepl.com/de/pro#developer) a (free) DeepL.com developer account to get an "Auth-Key" for the translate api.

## Examples

### Translation
Converts a Wikipedia article to markdown and translate it
```shell
$ export W2D_DEEPL_AUTH_KEY=aaaa-bbb-ccc

# Convert german article to markdown and translates it to russian. Output is always printed to stdout.
$ w2d translate ru https://de.wikipedia.org/wiki/Warentrenner > warentrenner_ru.md

# You can also pass a source-language in case auto-detection by DeepL.com fails
$ w2d translate -s nl ru https://nl.wikipedia.org/wiki/Beurtbalkje

# Convert and translate an article (html) stored on disk
$ w2d translate ru - < Warentrenner.html > warentrenner_ru.md

# Translate article fetched with wget to italian and print the result to the terminal
$ wget -q https://en.wikipedia.org/wiki/Hearth -O - | ./w2d translate it -
```

### Convert only
Convert articles to markdown without translating. No DeepL API-Key is required for these use-cases.

#### Read Wikipedia in your terminal
```shell
# Use glow to render and beautify markdown in your terminal
$ w2d markdown https://en.wikipedia.org/wiki/Hearth | glow -p 

# Or simply use less..
$ w2d markdown https://en.wikipedia.org/wiki/Hearth | less -r 
```

#### Work with articles on disk
```shell
# Convert and store an article as markdown
$ w2d markdown https://en.wikipedia.org/wiki/Hearth > hearth_en.md

# Convert article (html) stored on disk
$ w2d markdown - < Warentrenner.html > warentrenner_ru.md
```

### Misc

List source and target languages supported by the DeepL.com api:
```shell
$ export W2D_DEEPL_AUTH_KEY=aaaa-bbb-ccc

# List source languages 
$ w2d list-languages -t source

BG - Bulgarian (formality_support: false)
CS - Czech (formality_support: false)
DA - Danish (formality_support: false)
DE - German (formality_support: false)
EL - Greek (formality_support: false)
EN - English (formality_support: false)
ES - Spanish (formality_support: false)
ET - Estonian (formality_support: false)
FI - Finnish (formality_support: false)
FR - French (formality_support: false)
HU - Hungarian (formality_support: false)
IT - Italian (formality_support: false)
JA - Japanese (formality_support: false)
LT - Lithuanian (formality_support: false)
LV - Latvian (formality_support: false)
NL - Dutch (formality_support: false)
PL - Polish (formality_support: false)
PT - Portuguese (formality_support: false)
RO - Romanian (formality_support: false)
RU - Russian (formality_support: false)
SK - Slovak (formality_support: false)
SL - Slovenian (formality_support: false)
SV - Swedish (formality_support: false)
ZH - Chinese (formality_support: false)

# List target languages
$ w2d list-languages -t target

BG - Bulgarian (formality_support: false)
CS - Czech (formality_support: false)
DA - Danish (formality_support: false)
DE - German (formality_support: false)
EL - Greek (formality_support: false)
EN - English (formality_support: false)
ES - Spanish (formality_support: false)
ET - Estonian (formality_support: false)
FI - Finnish (formality_support: false)
FR - French (formality_support: false)
HU - Hungarian (formality_support: false)
IT - Italian (formality_support: false)
JA - Japanese (formality_support: false)
LT - Lithuanian (formality_support: false)
LV - Latvian (formality_support: false)
NL - Dutch (formality_support: false)
PL - Polish (formality_support: false)
PT - Portuguese (formality_support: false)
RO - Romanian (formality_support: false)
RU - Russian (formality_support: false)
SK - Slovak (formality_support: false)
SL - Slovenian (formality_support: false)
SV - Swedish (formality_support: false)
ZH - Chinese (formality_support: false)
```
