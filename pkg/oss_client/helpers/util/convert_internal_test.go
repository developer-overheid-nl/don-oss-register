package util

import (
	"os"
	"path/filepath"
	"testing"
)

func TestDonCheckerPublicCodeValidatorUsesLatestPublicCodeStandardForVersion070(t *testing.T) {
	tempDir := t.TempDir()
	argsFile := filepath.Join(tempDir, "npx-args.txt")
	fakeNpx := filepath.Join(tempDir, "npx")

	script := `#!/bin/sh
set -eu
printf '%s\n' "$@" > "$NPX_ARGS_FILE"

if [ "$#" -ne 7 ] ||
  [ "$1" != "--yes" ] ||
  [ "$2" != "@developer-overheid-nl/don-checker@latest" ] ||
  [ "$3" != "validate" ] ||
  [ "$4" != "--standard" ] ||
  [ "$5" != "publiccode" ] ||
  [ "$6" != "--input" ]; then
  echo "unexpected npx arguments: $*" >&2
  exit 42
fi

/usr/bin/grep -q 'publiccodeYmlVersion: "0.7.0"' "$7"
/usr/bin/grep -q 'softwareType: addon' "$7"
`

	if err := os.WriteFile(fakeNpx, []byte(script), 0o755); err != nil {
		t.Fatalf("write fake npx: %v", err)
	}

	t.Setenv("PATH", tempDir)
	t.Setenv("NPX_ARGS_FILE", argsFile)

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
