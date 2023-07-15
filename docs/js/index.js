requirejs.config({
    "baseUrl": "js/lib",
    "paths": {
        "jquery": "//code.jquery.com/jquery-3.7.0.min",
        "jquery-ui": "//code.jquery.com/ui/1.13.2/jquery-ui.min",
        "notie": "//unpkg.com/notie@4.3.1/dist/notie.min",
        "msgpack": "//unpkg.com/@msgpack/msgpack@3.0.0-beta2/dist.es5+umd/msgpack.min",
    }
});

define(['jquery', 'jquery-ui', 'notie', 'msgpack'], ($, ui, notie, msgpack) => {
    loadNames($, msgpack, notie);

    $(() => {
        const input = $("#search");
        input.focus();
    });
});

function search(query, max) {
    query = query.toLowerCase();
    const results = []
    for (const [group, m] of Object.entries(window.snos)) {
        for (const [id, name] of Object.entries(m)) {
            if (name.toLowerCase().includes(query)) {
                results.push({
                    label: `[${SnoGroups[group]}] ${name}`,
                    group,
                    name,
                    id,
                });
                if (results.length === max) {
                    return results
                }
            }
        }
    }
    return results
}

function entryName(item) {
    return `[${SnoGroups[item.group]}] ${item.name}`
}

function loadNames($, msgpack, notie) {
    const req = new XMLHttpRequest();
    req.open("GET", "names.mpk", true);
    req.responseType = "arraybuffer";
    req.onload = function () {
        $(() => {
            window.snos = msgpack.decode(req.response);
            $("#search").autocomplete({
                source: function(request, response) {
                    response(search(request.term, 100))
                },
                select: function(event, ui) {
                    const url = `sno/${ui.item.id}.html`
                    $.get({
                        url: `sno/${ui.item.id}.html`,
                        cache: true,
                    }).done(() => {
                        $(location).prop('href', url);
                    }).fail(() => {
                        notie.alert({
                            type: 'error',
                            text: 'SNO not found!',
                            buttonText: 'Moo!',
                            time: 100,
                        });
                        $("#search").focus();
                    })
                },
                focus: function (event, ui) {
                    this.value = entryName(ui.item);
                    event.preventDefault();
                },
            })
        })
    };
    req.send();
}

const SnoGroups = {
    "-3": "Unknown",
    "-2": "Code",
    "-1": "None",
    "1": "Actor",
    "2": "NPCComponentSet",
    "3": "AIBehavior",
    "4": "AIState",
    "5": "AmbientSound",
    "6": "Anim",
    "7": "Anim2D",
    "8": "AnimSet",
    "9": "Appearance",
    "10": "Hero",
    "11": "Cloth",
    "12": "Conversation",
    "13": "ConversationList",
    "14": "EffectGroup",
    "15": "Encounter",
    "17": "Explosion",
    "18": "FlagSet",
    "19": "Font",
    "20": "GameBalance",
    "21": "Global",
    "22": "LevelArea",
    "23": "Light",
    "24": "MarkerSet",
    "26": "Observer",
    "27": "Particle",
    "28": "Physics",
    "29": "Power",
    "31": "Quest",
    "32": "Rope",
    "33": "Scene",
    "35": "Script",
    "36": "ShaderMap",
    "37": "Shader",
    "38": "Shake",
    "39": "SkillKit",
    "40": "Sound",
    "42": "StringList",
    "43": "Surface",
    "44": "Texture",
    "45": "Trail",
    "46": "UI",
    "47": "Weather",
    "48": "World",
    "49": "Recipe",
    "51": "Condition",
    "52": "TreasureClass",
    "53": "Account",
    "57": "Material",
    "59": "Lore",
    "60": "Reverb",
    "62": "Music",
    "63": "Tutorial",
    "67": "AnimTree",
    "68": "Vibration",
    "71": "wWiseSoundBank",
    "72": "Speaker",
    "73": "Item",
    "74": "PlayerClass",
    "76": "FogVolume",
    "77": "Biome",
    "78": "Wall",
    "79": "SoundTable",
    "80": "Subzone",
    "81": "MaterialValue",
    "82": "MonsterFamily",
    "83": "TileSet",
    "84": "Population",
    "85": "MaterialValueSet",
    "86": "WorldState",
    "87": "Schedule",
    "88": "VectorField",
    "90": "Storyboard",
    "92": "Territory",
    "93": "AudioContext",
    "94": "VOProcess",
    "95": "DemonScroll",
    "96": "QuestChain",
    "97": "LoudnessPreset",
    "98": "ItemType",
    "99": "Achievement",
    "100": "Crafter",
    "101": "HoudiniParticlesSim",
    "102": "Movie",
    "103": "TiledStyle",
    "104": "Affix",
    "105": "Reputation",
    "106": "ParagonNode",
    "107": "MonsterAffix",
    "108": "ParagonBoard",
    "109": "SetItemBonus",
    "110": "StoreProduct",
    "111": "ParagonGlyph",
    "112": "ParagonGlyphAffix",
    "114": "Challenge",
    "115": "MarkingShape",
    "116": "ItemRequirement",
    "117": "Boost",
    "118": "Emote",
    "119": "Jewelry",
    "120": "PlayerTitle",
    "121": "Emblem",
    "122": "Dye",
    "123": "FogOfWar",
    "124": "ParagonThreshold",
    "125": "AIAwareness",
    "126": "TrackedReward",
    "127": "CollisionSettings",
    "128": "Aspect",
    "129": "ABTest",
    "130": "Stagger",
    "131": "EyeColor",
    "132": "Makeup",
    "133": "MarkingColor",
    "134": "HairColor",
    "135": "DungeonAffix",
    "136": "Activity",
    "138": "HairStyle",
    "139": "FacialHair",
    "140": "Face",
}