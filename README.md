# RRL Results Formatter

Ein Formatierer für Simracing-Ligen.

## Datenformat

### Rennergebnis `results.csv`

```csv
1. Zeile: URL-Link zu Rennveranstaltung
2. Zeile: Fahrer-ID Schnellste Rennrunde
3. Zeile: Fahrer auf Pos 1
4. Zeile: Fahrer auf Pos 2
...
Vorletzte Zeile: Fahrer auf letzter Position
Letzte Zeile: Position des ersten Fahrers mit DNF, 0 wenn Rennen ohne DNF
```
### Serieneinteilung: `einteilung.csv`

```csv
Serie,Fahrer-Id
```

### Teams: `teams.csv`

```csv
Team-Name,Fahrer-Id Fahrer 1, Fahrer-Id Fahrer 2
```
