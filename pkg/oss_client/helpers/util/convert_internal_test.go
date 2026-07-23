package util

import "testing"

func TestDonCheckerPublicCodeValidatorUsesLatestPublicCodeStandardForVersion070(t *testing.T) {
	input := `publiccodeYmlVersion: "0.7.0"
name: "Agent Skills: Open Source Repository"
url: https://github.com/developer-overheid-nl/skills-open-source-repo
softwareType: addon
platforms:
  - linux
  - mac
  - windows
developmentStatus: stable
description:
  nl:
    shortDescription: >-
      Een Agent Skill die je helpt bij het aanmaken van alle benodigde bestanden
      voor een goed open source project.
    longDescription: |
      De Open Source Repo Skill is een Agent Skill die ontwikkelaars bij de
      Nederlandse overheid helpt om hun repository klaar te maken voor open source
      publicatie. De skill genereert bestanden zoals publiccode.yml, README.md,
      CONTRIBUTING.md, SECURITY.md en LICENSE op basis van projectinformatie.
    features:
      - Genereert publiccode.yml op basis van projectinformatie
legal:
  license: EUPL-1.2
  mainCopyrightOwner: developer.overheid.nl
localisation:
  availableLanguages:
    - nl
  localisationReady: false
maintenance:
  type: internal
  contacts:
    - name: developer.overheid.nl
      email: developer.overheid@geonovum.nl
      affiliation: Geonovum
`

	if err := (donCheckerPublicCodeValidator{}).ValidatePublicCode(input); err != nil {
		t.Fatalf("validate publiccode 0.7.0: %v", err)
	}
}
