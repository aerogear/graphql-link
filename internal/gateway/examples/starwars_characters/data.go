package starwars_characters

const data = `
humans:
- id:        "1000"
  name:      "Luke Skywalker"
  friendIds:   ["1002", "1003", "2000", "2001"]
  appearsIn: ["NEWHOPE", "EMPIRE", "JEDI"]
  height:    1.72
  mass:      77
  starships: ["3001", "3003"]
- id:        "1001"
  name:      "Darth Vader"
  friendIds:   ["1004"]
  appearsIn: ["NEWHOPE", "EMPIRE", "JEDI"]
  height:    2.02
  mass:      136
  starships: ["3002"]
- id:        "1002"
  name:      "Han Solo"
  friendIds:   ["1000", "1003", "2001"]
  appearsIn: ["NEWHOPE", "EMPIRE", "JEDI"]
  height:    1.8
  mass:      80
  starships: ["3000", "3003"]
- id:        "1003"
  name:      "Leia Organa"
  friendIds:   ["1000", "1002", "2000", "2001"]
  appearsIn: ["NEWHOPE", "EMPIRE", "JEDI"]
  height:    1.5
  mass:      49
- id:        "1004"
  name:      "Wilhuff Tarkin"
  friendIds:   ["1001"]
  appearsIn: ["NEWHOPE"]
  height:    1.8
  mass:      0

droids:
- id:              "2000"
  name:            "C-3PO"
  friendIds:         ["1000", "1002", "1003", "2001"]
  appearsIn:       ["NEWHOPE", "EMPIRE", "JEDI"]
  primaryFunction: "Protocol"
- id:              "2001"
  name:            "R2-D2"
  friendIds:         ["1000", "1002", "1003"]
  appearsIn:       ["NEWHOPE", "EMPIRE", "JEDI"]
  primaryFunction: "Astromech"
`