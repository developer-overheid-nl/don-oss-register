package models

// SoftwareTypeLabels bevat de labels en omschrijvingen per softwareType waarde.
// Gebaseerd op https://yml.publiccode.tools/schema.core.html
var SoftwareTypeLabels = map[string][2]string{
	"standalone/web":     {"Web applicatie", "Software toegankelijk via een webbrowser."},
	"standalone/desktop": {"Desktop applicatie", "Software die lokaal op een desktop draait."},
	"standalone/mobile":  {"Mobiele applicatie", "Software die op een mobiel apparaat draait."},
	"standalone/backend": {"Backend / API", "Server-side software of API."},
	"standalone/iot":     {"IoT", "Software voor Internet of Things apparaten."},
	"standalone/other":   {"Overig standalone", "Standalone software die niet in een andere categorie valt."},
	"addon":              {"Addon / Plugin", "Uitbreiding voor bestaande software."},
	"library":            {"Library", "Herbruikbare bibliotheek voor ontwikkelaars."},
	"configurationFiles": {"Configuratiebestanden", "Configuraties of templates voor andere software."},
}

// DevelopmentStatusLabels bevat de labels en omschrijvingen per developmentStatus waarde.
var DevelopmentStatusLabels = map[string][2]string{
	"concept":     {"Concept", "Software in een vroeg conceptstadium."},
	"development": {"In ontwikkeling", "Software die actief wordt ontwikkeld."},
	"beta":        {"Beta", "Software in de beta-testfase."},
	"stable":      {"Stabiel", "Productieklare, stabiele software."},
	"obsolete":    {"Verouderd", "Software die niet meer actief onderhouden wordt."},
}

// MaintenanceTypeLabels bevat de labels en omschrijvingen per maintenance type waarde.
var MaintenanceTypeLabels = map[string][2]string{
	"none":      {"Geen onderhoud", "Er is geen actief onderhoud."},
	"internal":  {"Intern", "Onderhouden door de eigen organisatie."},
	"contract":  {"Contract", "Onderhoud via een externe contractpartij."},
	"community": {"Community", "Onderhouden door een open source community."},
}

// PlatformLabels bevat de labels en omschrijvingen per platform waarde.
var PlatformLabels = map[string][2]string{
	"web":     {"Web", "Toegankelijk via een webbrowser."},
	"windows": {"Windows", "Beschikbaar voor Microsoft Windows."},
	"mac":     {"macOS", "Beschikbaar voor Apple macOS."},
	"linux":   {"Linux", "Beschikbaar voor Linux."},
	"ios":     {"iOS", "Beschikbaar voor Apple iOS."},
	"android": {"Android", "Beschikbaar voor Android."},
}

// LanguageLabels bevat de Nederlandse namen per ISO 639-1 taalcode.
var LanguageLabels = map[string]string{
	"nl": "Nederlands",
	"en": "Engels",
	"de": "Duits",
	"fr": "Frans",
	"es": "Spaans",
	"pt": "Portugees",
	"it": "Italiaans",
	"pl": "Pools",
	"ar": "Arabisch",
}
