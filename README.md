# RRL Results Formatter

Ein Formatierer für Simracing-Ligen.

## Datenformat

### Rennergebnis `results.csv`

```csv
1. Zeile: URL-Link zu Rennveranstaltung
2. Zeile: Position des ersten Fahrers mit DNF, 0 wenn Rennen ohne DNF
3. Zeile: Fahrer-ID Schnellste Rennrunde
4. Zeile: Fahrer auf Pos 1
5. Zeile: Fahrer auf Pos 2
...
Letzte Zeile: Fahrer auf letzter Position
```
### Serieneinteilung: `einteilung.csv`

```csv
Serie,Fahrer-Id,Klasse
```

### Teams: `teams.csv`

```csv
Team-Name,Fahrer-Id Fahrer 1, Fahrer-Id Fahrer 2,Klasse
```
