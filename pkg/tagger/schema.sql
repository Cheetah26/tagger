PRAGMA auto_vacuum = FULL;
PRAGMA foreign_keys = ON;

-- Data tables
CREATE TABLE IF NOT EXISTS Tags (
  Id INTEGER PRIMARY KEY,
  Name TEXT NOT NULL
);

CREATE TABLE IF NOT EXISTS Files (
  Id INTEGER PRIMARY KEY,
  Hash TEXT NOT NULL UNIQUE,
  Filetype TEXT NOT NULL,
  Name TEXT,
  Description TEXT
);

-- Relationship tables
CREATE TABLE IF NOT EXISTS FileTag (
  FileId INTEGER REFERENCES Files(Id) ON DELETE CASCADE,
  TagId INTEGER REFERENCES Tags(Id) ON DELETE CASCADE,
  PRIMARY KEY (FileId, TagId)
);

CREATE TABLE IF NOT EXISTS TagTag (
  ParentTagId INTEGER REFERENCES Tags(Id) ON DELETE CASCADE,
  ChildTagId INTEGER REFERENCES Tags(Id) ON DELETE CASCADE,
  PRIMARY KEY (ParentTagId, ChildTagId),
  CHECK(ParentTagId != ChildTagId)
);