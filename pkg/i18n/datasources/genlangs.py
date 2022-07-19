import json
from string import Template
from pprint import pprint as _p

with open('language-codes.json', 'r') as langcodes_fp:
    lang_codes = json.load(langcodes_fp)

lang_names = {}
for lce in lang_codes:
    lang_names[lce['alpha2']] = lce['English']


with open('country-list.json', 'r') as countrylist_fp:
    country_list_json = json.load(countrylist_fp)

country_list = {}
for cle in country_list_json:
    country_list[cle['Code']] = cle['Name']

with open('ietf-language-tags.json', 'r') as langs_fp:
    ietf_langs = json.load(langs_fp)



langs = {}
for ile in ietf_langs:
    lt = ile['langType']
    t = ile['territory']
    l = ile['lang']

    if lt not in langs:
        ln = None
        if lt in lang_names:
            ln = lang_names[lt]
        langs[lt] = {
            'name': ln,
            'variants': {},
        }
    tname = ""
    if t in country_list:
        tname = country_list[t]
    if t is not None:
        langs[lt]['variants'][t] = tname
    # print(f"{lt} _ {t} ({tname}) _ {l}")

# _p(langs)

lang_tmpl = Template("""
    Lang{
        Name: "$name",
        Code: "$code",
        Variants: []LangVariant{ $lang_variants
        },
    },""")

variant_tmpl = Template("""
            LangVariant{
                Name: "$variant_name",
                Code: "$variant_code",
            },""")


hardcoded_langs = "var hardcodedLangs = []Lang{"
hardcoded_langs_idx = "var hardcodedLangIdx = map[string]int{"

for idx, code in enumerate(sorted(langs.keys())):
    langdef = langs[code]
    variant_gen = ""
    for varcode, varname in langdef['variants'].items():
        if varname is None:
            varname = ""
        vg = variant_tmpl.substitute(
                variant_name=varname,
                variant_code=varcode)
        variant_gen = variant_gen + vg
    lang_name = langdef['name'] or ""
    lang_gen = lang_tmpl.substitute(
            name=lang_name,
            code=code,
            lang_variants=variant_gen
    )
    hardcoded_langs = hardcoded_langs + lang_gen
    hardcoded_langs_idx += f"\n\"{code}\": {idx},"


hardcoded_langs = hardcoded_langs + "}"
hardcoded_langs_idx = hardcoded_langs_idx + "}"

output = "package langs\n\n" + hardcoded_langs + "\n\n" + hardcoded_langs_idx

print(output)


with open('hclangs.json', 'w') as w:
    json.dump(langs, w)

