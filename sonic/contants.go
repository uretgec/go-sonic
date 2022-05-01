package sonic

// Sonic Channel Modes
const (
	ChannelUninitialized = "uninitialized"
	ChannelSearch        = "search"
	ChannelIngest        = "ingest"
	ChannelControl       = "control"
)

// Sonic Commands
const (
	CmdPing = "PING" // PING
	CmdQuit = "QUIT" // QUIT

	// ChannelUninitialized
	CmdSearchStart = "START"

	// Search Mode: ChannelSearch
	CmdSearchQuery   = "QUERY"
	CmdSearchSuggest = "SUGGEST"
	CmdSearchPing    = "PING" // PING

	// Ingest Mode: ChannelIngest
	CmdIngestPush   = "PUSH"
	CmdIngestPop    = "POP"
	CmdIngestCount  = "COUNT"
	CmdIngestFlushc = "FLUSHC"
	CmdIngestFlushb = "FLUSHB"
	CmdIngestFlusho = "FLUSHO"

	// Control Mode: ChannelControl - Use only for administration
	CmdControlTrigger = "TRIGGER"
	CmdControlInfo    = "INFO"
)

// Trigger Action
const (
	TriggerActionConsolidate = "consolidate"
	TriggerActionBackup      = "backup"
	TriggerActionRestore     = "restore"
)

// Trigger Data
const (
	TriggerDataBackup  = "backup"
	TriggerDataRestore = "restore"
)

// Sonic Supported Language: an ISO 639-3 locale code
const (
	LangAutoDetect = ""
	LangNone       = "none"
	LangAfr        = "afr"
	LangAka        = "aka"
	LangAmh        = "amh"
	LangAra        = "ara"
	LangAzj        = "azj"
	LangBel        = "bel"
	LangBen        = "ben"
	LangBho        = "bho"
	LangBul        = "bul"
	LangCat        = "cat"
	LangCeb        = "ceb"
	LangCes        = "ces"
	LangCmn        = "cmn"
	LangDan        = "dan"
	LangDeu        = "deu"
	LangEll        = "ell"
	LangEng        = "eng"
	LangEpo        = "epo"
	LangEst        = "est"
	LangFin        = "fin"
	LangFra        = "fra"
	LangGuj        = "guj"
	LangHat        = "hat"
	LangHau        = "hau"
	LangHeb        = "heb"
	LangHin        = "hin"
	LangHrv        = "hrv"
	LangHun        = "hun"
	LangIbo        = "ibo"
	LangIlo        = "ilo"
	LangInd        = "ind"
	LangIta        = "ita"
	LangJav        = "jav"
	LangJpn        = "jpn"
	LangKan        = "kan"
	LangKat        = "kat"
	LangKhm        = "khm"
	LangKin        = "kin"
	LangKor        = "kor"
	LangKur        = "kur"
	LangLat        = "lat"
	LangLav        = "lav"
	LangLit        = "lit"
	LangMai        = "mai"
	LangMal        = "mal"
	LangMar        = "mar"
	LangMkd        = "mkd"
	LangMlg        = "mlg"
	LangMod        = "mod"
	LangMya        = "mya"
	LangNep        = "nep"
	LangNld        = "nld"
	LangNno        = "nno"
	LangNob        = "nob"
	LangNya        = "nya"
	LangOri        = "ori"
	LangOrm        = "orm"
	LangPan        = "pan"
	LangPes        = "pes"
	LangPol        = "pol"
	LangPor        = "por"
	LangRon        = "ron"
	LangRun        = "run"
	LangRus        = "rus"
	LangSin        = "sin"
	LangSkr        = "skr"
	LangSlk        = "slk"
	LangSlv        = "slv"
	LangSna        = "sna"
	LangSom        = "som"
	LangSpa        = "spa"
	LangSrp        = "srp"
	LangSwe        = "swe"
	LangTam        = "tam"
	LangTel        = "tel"
	LangTgl        = "tgl"
	LangTha        = "tha"
	LangTir        = "tir"
	LangTuk        = "tuk"
	LangTur        = "tur"
	LangUig        = "uig"
	LangUkr        = "ukr"
	LangUrd        = "urd"
	LangUzb        = "uzb"
	LangVie        = "vie"
	LangYdd        = "ydd"
	LangYor        = "yor"
	LangZul        = "zul"
)
