package d4

type SnoGroup int32

const (
	SnoGroupUnknown             SnoGroup = -3
	SnoGroupCode                SnoGroup = -2
	SnoGroupNone                SnoGroup = -1
	SnoGroupActor               SnoGroup = 1
	SnoGroupNPCComponentSet     SnoGroup = 2
	SnoGroupAIBehavior          SnoGroup = 3
	SnoGroupAIState             SnoGroup = 4
	SnoGroupAmbientSound        SnoGroup = 5
	SnoGroupAnim                SnoGroup = 6
	SnoGroupAnim2D              SnoGroup = 7
	SnoGroupAnimSet             SnoGroup = 8
	SnoGroupAppearance          SnoGroup = 9
	SnoGroupHero                SnoGroup = 10
	SnoGroupCloth               SnoGroup = 11
	SnoGroupConversation        SnoGroup = 12
	SnoGroupConversationList    SnoGroup = 13
	SnoGroupEffectGroup         SnoGroup = 14
	SnoGroupEncounter           SnoGroup = 15
	SnoGroupExplosion           SnoGroup = 17
	SnoGroupFlagSet             SnoGroup = 18
	SnoGroupFont                SnoGroup = 19
	SnoGroupGameBalance         SnoGroup = 20
	SnoGroupGlobal              SnoGroup = 21
	SnoGroupLevelArea           SnoGroup = 22
	SnoGroupLight               SnoGroup = 23
	SnoGroupMarkerSet           SnoGroup = 24
	SnoGroupObserver            SnoGroup = 26
	SnoGroupParticle            SnoGroup = 27
	SnoGroupPhysics             SnoGroup = 28
	SnoGroupPower               SnoGroup = 29
	SnoGroupQuest               SnoGroup = 31
	SnoGroupRope                SnoGroup = 32
	SnoGroupScene               SnoGroup = 33
	SnoGroupScript              SnoGroup = 35
	SnoGroupShaderMap           SnoGroup = 36
	SnoGroupShader              SnoGroup = 37
	SnoGroupShake               SnoGroup = 38
	SnoGroupSkillKit            SnoGroup = 39
	SnoGroupSound               SnoGroup = 40
	SnoGroupStringList          SnoGroup = 42
	SnoGroupSurface             SnoGroup = 43
	SnoGroupTexture             SnoGroup = 44
	SnoGroupTrail               SnoGroup = 45
	SnoGroupUI                  SnoGroup = 46
	SnoGroupWeather             SnoGroup = 47
	SnoGroupWorld               SnoGroup = 48
	SnoGroupRecipe              SnoGroup = 49
	SnoGroupCondition           SnoGroup = 51
	SnoGroupTreasureClass       SnoGroup = 52
	SnoGroupAccount             SnoGroup = 53
	SnoGroupMaterial            SnoGroup = 57
	SnoGroupLore                SnoGroup = 59
	SnoGroupReverb              SnoGroup = 60
	SnoGroupMusic               SnoGroup = 62
	SnoGroupTutorial            SnoGroup = 63
	SnoGroupAnimTree            SnoGroup = 67
	SnoGroupVibration           SnoGroup = 68
	SnoGroupwWiseSoundBank      SnoGroup = 71
	SnoGroupSpeaker             SnoGroup = 72
	SnoGroupItem                SnoGroup = 73
	SnoGroupPlayerClass         SnoGroup = 74
	SnoGroupFogVolume           SnoGroup = 76
	SnoGroupBiome               SnoGroup = 77
	SnoGroupWall                SnoGroup = 78
	SnoGroupSoundTable          SnoGroup = 79
	SnoGroupSubzone             SnoGroup = 80
	SnoGroupMaterialValue       SnoGroup = 81
	SnoGroupMonsterFamily       SnoGroup = 82
	SnoGroupTileSet             SnoGroup = 83
	SnoGroupPopulation          SnoGroup = 84
	SnoGroupMaterialValueSet    SnoGroup = 85
	SnoGroupWorldState          SnoGroup = 86
	SnoGroupSchedule            SnoGroup = 87
	SnoGroupVectorField         SnoGroup = 88
	SnoGroupStoryboard          SnoGroup = 90
	SnoGroupTerritory           SnoGroup = 92
	SnoGroupAudioContext        SnoGroup = 93
	SnoGroupVOProcess           SnoGroup = 94
	SnoGroupDemonScroll         SnoGroup = 95
	SnoGroupQuestChain          SnoGroup = 96
	SnoGroupLoudnessPreset      SnoGroup = 97
	SnoGroupItemType            SnoGroup = 98
	SnoGroupAchievement         SnoGroup = 99
	SnoGroupCrafter             SnoGroup = 100
	SnoGroupHoudiniParticlesSim SnoGroup = 101
	SnoGroupMovie               SnoGroup = 102
	SnoGroupTiledStyle          SnoGroup = 103
	SnoGroupAffix               SnoGroup = 104
	SnoGroupReputation          SnoGroup = 105
	SnoGroupParagonNode         SnoGroup = 106
	SnoGroupMonsterAffix        SnoGroup = 107
	SnoGroupParagonBoard        SnoGroup = 108
	SnoGroupSetItemBonus        SnoGroup = 109
	SnoGroupStoreProduct        SnoGroup = 110
	SnoGroupParagonGlyph        SnoGroup = 111
	SnoGroupParagonGlyphAffix   SnoGroup = 112
	SnoGroupChallenge           SnoGroup = 114
	SnoGroupMarkingShape        SnoGroup = 115
	SnoGroupItemRequirement     SnoGroup = 116
	SnoGroupBoost               SnoGroup = 117
	SnoGroupEmote               SnoGroup = 118
	SnoGroupJewelry             SnoGroup = 119
	SnoGroupPlayerTitle         SnoGroup = 120
	SnoGroupEmblem              SnoGroup = 121
	SnoGroupDye                 SnoGroup = 122
	SnoGroupFogOfWar            SnoGroup = 123
	SnoGroupParagonThreshold    SnoGroup = 124
	SnoGroupAIAwareness         SnoGroup = 125
	SnoGroupTrackedReward       SnoGroup = 126
	SnoGroupCollisionSettings   SnoGroup = 127
	SnoGroupAspect              SnoGroup = 128
	SnoGroupABTest              SnoGroup = 129
	SnoGroupStagger             SnoGroup = 130
	SnoGroupEyeColor            SnoGroup = 131
	SnoGroupMakeup              SnoGroup = 132
	SnoGroupMarkingColor        SnoGroup = 133
	SnoGroupHairColor           SnoGroup = 134
	SnoGroupDungeonAffix        SnoGroup = 135
	SnoGroupActivity            SnoGroup = 136
	SnoSeason                   SnoGroup = 136
	SnoGroupHairStyle           SnoGroup = 138
	SnoGroupFacialHair          SnoGroup = 139
	SnoGroupFace                SnoGroup = 140
	SnoAiCoordinator            SnoGroup = 144
	MaxSnoGroups                         = 147
)

func (g SnoGroup) String() string {
	switch g {
	case -2:
		return "Code"
	case -1:
		return "None"
	case 1:
		return "Actor"
	case 2:
		return "NPCComponentSet"
	case 3:
		return "AIBehavior"
	case 4:
		return "AIState"
	case 5:
		return "AmbientSound"
	case 6:
		return "Anim"
	case 7:
		return "Anim2D"
	case 8:
		return "AnimSet"
	case 9:
		return "Appearance"
	case 10:
		return "Hero"
	case 11:
		return "Cloth"
	case 12:
		return "Conversation"
	case 13:
		return "ConversationList"
	case 14:
		return "EffectGroup"
	case 15:
		return "Encounter"
	case 17:
		return "Explosion"
	case 18:
		return "FlagSet"
	case 19:
		return "Font"
	case 20:
		return "GameBalance"
	case 21:
		return "Global"
	case 22:
		return "LevelArea"
	case 23:
		return "Light"
	case 24:
		return "MarkerSet"
	case 26:
		return "Observer"
	case 27:
		return "Particle"
	case 28:
		return "Physics"
	case 29:
		return "Power"
	case 31:
		return "Quest"
	case 32:
		return "Rope"
	case 33:
		return "Scene"
	case 35:
		return "Script"
	case 36:
		return "ShaderMap"
	case 37:
		return "Shader"
	case 38:
		return "Shake"
	case 39:
		return "SkillKit"
	case 40:
		return "Sound"
	case 42:
		return "StringList"
	case 43:
		return "Surface"
	case 44:
		return "Texture"
	case 45:
		return "Trail"
	case 46:
		return "UI"
	case 47:
		return "Weather"
	case 48:
		return "World"
	case 49:
		return "Recipe"
	case 51:
		return "Condition"
	case 52:
		return "TreasureClass"
	case 53:
		return "Account"
	case 57:
		return "Material"
	case 59:
		return "Lore"
	case 60:
		return "Reverb"
	case 62:
		return "Music"
	case 63:
		return "Tutorial"
	case 67:
		return "AnimTree"
	case 68:
		return "Vibration"
	case 71:
		return "wWiseSoundBank"
	case 72:
		return "Speaker"
	case 73:
		return "Item"
	case 74:
		return "PlayerClass"
	case 76:
		return "FogVolume"
	case 77:
		return "Biome"
	case 78:
		return "Wall"
	case 79:
		return "SoundTable"
	case 80:
		return "Subzone"
	case 81:
		return "MaterialValue"
	case 82:
		return "MonsterFamily"
	case 83:
		return "TileSet"
	case 84:
		return "Population"
	case 85:
		return "MaterialValueSet"
	case 86:
		return "WorldState"
	case 87:
		return "Schedule"
	case 88:
		return "VectorField"
	case 90:
		return "Storyboard"
	case 92:
		return "Territory"
	case 93:
		return "AudioContext"
	case 94:
		return "VOProcess"
	case 95:
		return "DemonScroll"
	case 96:
		return "QuestChain"
	case 97:
		return "LoudnessPreset"
	case 98:
		return "ItemType"
	case 99:
		return "Achievement"
	case 100:
		return "Crafter"
	case 101:
		return "HoudiniParticlesSim"
	case 102:
		return "Movie"
	case 103:
		return "TiledStyle"
	case 104:
		return "Affix"
	case 105:
		return "Reputation"
	case 106:
		return "ParagonNode"
	case 107:
		return "MonsterAffix"
	case 108:
		return "ParagonBoard"
	case 109:
		return "SetItemBonus"
	case 110:
		return "StoreProduct"
	case 111:
		return "ParagonGlyph"
	case 112:
		return "ParagonGlyphAffix"
	case 114:
		return "Challenge"
	case 115:
		return "MarkingShape"
	case 116:
		return "ItemRequirement"
	case 117:
		return "Boost"
	case 118:
		return "Emote"
	case 119:
		return "Jewelry"
	case 120:
		return "PlayerTitle"
	case 121:
		return "Emblem"
	case 122:
		return "Dye"
	case 123:
		return "FogOfWar"
	case 124:
		return "ParagonThreshold"
	case 125:
		return "AIAwareness"
	case 126:
		return "TrackedReward"
	case 127:
		return "CollisionSettings"
	case 128:
		return "Aspect"
	case 129:
		return "ABTest"
	case 130:
		return "Stagger"
	case 131:
		return "EyeColor"
	case 132:
		return "Makeup"
	case 133:
		return "MarkingColor"
	case 134:
		return "HairColor"
	case 135:
		return "DungeonAffix"
	case 136:
		return "Activity"
	case 137:
		return "Season"
	case 138:
		return "HairStyle"
	case 139:
		return "FacialHair"
	case 140:
		return "Face"
	case 141:
		return "MercenaryClass"
	case 142:
		return "PassivePowerContainer"
	case 143:
		return "MountProfile"
	case 144:
		return "AICoordinator"
	case 145:
		return "CrafterTab"
	case 146:
		return "TownPortalCosmetic"
	case 147:
		return "AxeTest"
	case 148:
		return "Wizard"
	case 149:
		return "FootstepTable"
	case 150:
		return "Modal"
	case 151:
		return "CollectiblePower"
	case 152:
		return "AppearenceSet"
	case 153:
		return "Preset"
	default:
		return "Unknown"
	}
}

func (g SnoGroup) Ext() string {
	switch g {
	case 0:
		return ""
	case 1:
		return ".acr"
	case 2:
		return ".npc"
	case 3:
		return ".aib"
	case 4:
		return ".ais"
	case 5:
		return ".ams"
	case 6:
		return ".ani"
	case 7:
		return ".an2"
	case 8:
		return ".ans"
	case 9:
		return ".app"
	case 10:
		return ".hro"
	case 11:
		return ".clt"
	case 12:
		return ".cnv"
	case 13:
		return ".cnl"
	case 14:
		return ".efg"
	case 15:
		return ".enc"
	case 16:
		return ""
	case 17:
		return ".xpl"
	case 18:
		return ".flg"
	case 19:
		return ".fnt"
	case 20:
		return ".gam"
	case 21:
		return ".glo"
	case 22:
		return ".lvl"
	case 23:
		return ".lit"
	case 24:
		return ".mrk"
	case 25:
		return ""
	case 26:
		return ".obs"
	case 27:
		return ".prt"
	case 28:
		return ".phy"
	case 29:
		return ".pow"
	case 30:
		return ""
	case 31:
		return ".qst"
	case 32:
		return ".rop"
	case 33:
		return ".scn"
	case 34:
		return ""
	case 35:
		return ".scr"
	case 36:
		return ".shm"
	case 37:
		return ".shd"
	case 38:
		return ".shk"
	case 39:
		return ".skl"
	case 40:
		return ".snd"
	case 41:
		return ""
	case 42:
		return ".stl"
	case 43:
		return ".srf"
	case 44:
		return ".tex"
	case 45:
		return ".trl"
	case 46:
		return ".ui"
	case 47:
		return ".wth"
	case 48:
		return ".wrl"
	case 49:
		return ".rcp"
	case 50:
		return ""
	case 51:
		return ".cnd"
	case 52:
		return ".trs"
	case 53:
		return ".acc"
	case 54:
		return ""
	case 55:
		return ""
	case 56:
		return ""
	case 57:
		return ".mat"
	case 58:
		return ""
	case 59:
		return ".lor"
	case 60:
		return ".rev"
	case 61:
		return ""
	case 62:
		return ".mus"
	case 63:
		return ".tut"
	case 64:
		return ""
	case 65:
		return ""
	case 66:
		return ""
	case 67:
		return ".ant"
	case 68:
		return ".vib"
	case 69:
		return ""
	case 70:
		return ""
	case 71:
		return ".wsb"
	case 72:
		return ".spk"
	case 73:
		return ".itm"
	case 74:
		return ".pcl"
	case 75:
		return ""
	case 76:
		return ".fog"
	case 77:
		return ".bio"
	case 78:
		return ".wal"
	case 79:
		return ".sdt"
	case 80:
		return ".sbz"
	case 81:
		return ".mtv"
	case 82:
		return ".mfm"
	case 83:
		return ".tst"
	case 84:
		return ".pop"
	case 85:
		return ".mvs"
	case 86:
		return ".wst"
	case 87:
		return ".sch"
	case 88:
		return ".vfd"
	case 89:
		return ""
	case 90:
		return ".stb"
	case 91:
		return ""
	case 92:
		return ".ter"
	case 93:
		return ".auc"
	case 94:
		return ".vop"
	case 95:
		return ".dss"
	case 96:
		return ".qc"
	case 97:
		return ".lou"
	case 98:
		return ".itt"
	case 99:
		return ".ach"
	case 100:
		return ".crf"
	case 101:
		return ".hps"
	case 102:
		return ".vid"
	case 103:
		return ".tsl"
	case 104:
		return ".aff"
	case 105:
		return ".rep"
	case 106:
		return ".pgn"
	case 107:
		return ".maf"
	case 108:
		return ".pbd"
	case 109:
		return ".set"
	case 110:
		return ".prd"
	case 111:
		return ".gph"
	case 112:
		return ".gaf"
	case 113:
		return ""
	case 114:
		return ".cha"
	case 115:
		return ".msh"
	case 116:
		return ".irq"
	case 117:
		return ".bst"
	case 118:
		return ".emo"
	case 119:
		return ".jwl"
	case 120:
		return ".pt"
	case 121:
		return ".emb"
	case 122:
		return ".dye"
	case 123:
		return ".fow"
	case 124:
		return ".pth"
	case 125:
		return ".aia"
	case 126:
		return ".trd"
	case 127:
		return ".col"
	case 128:
		return ".asp"
	case 129:
		return ".abt"
	case 130:
		return ".stg"
	case 131:
		return ".eye"
	case 132:
		return ".mak"
	case 133:
		return ".mcl"
	case 134:
		return ".hcl"
	case 135:
		return ".dax"
	case 136:
		return ".act"
	case 137:
		return ".sea"
	case 138:
		return ".har"
	case 139:
		return ".fhr"
	case 140:
		return ".fac"
	case 141:
		return ".mrc"
	case 142:
		return ".ppc"
	case 143:
		return ".mpp"
	case 144:
		return ".aic"
	case 145:
		return ".ctb"
	case 146:
		return ".tpc"
	case 147:
		return ".axe"
	case 148:
		return ".wiz"
	case 149:
		return ".fst"
	case 150:
		return ".mdl"
	case 151:
		return ".cpw"
	case 152:
		return ".aps"
	case 153:
		return ".pst"
	default:
		return ""
	}
}
