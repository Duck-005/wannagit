# wannagit  

wannagit is a **toy version of Git** written in **Go**.  
It is not intended for production use, but rather as a learning project to explore,
- How Git works internally (objects, commits, trees, index, refs, etc.)
- learn go as I go along instead of endless exercises.
- How to build a CLI tool in go using [Cobra](https://github.com/spf13/cobra).  

The code is riddled with bugs and random panics everywhere with error handling all over the place, so good luck! if you decide to contribute or use it anywhere else.

## Installation  

Clone the repository and build the binary:  

```bash
git clone https://github.com/Duck-005/wannagit.git
cd wannagit
go build -o wannagit
```

Then run wannagit commands from the project directory:
```bash
./wannagit init myrepo
```

## Commands

Wannagit supports the following commands:

#### `init`  
Initialize a new wannagit repo.  
```bash
wannagit init <path>
``` 

---

#### add
Add files to the staging area or the index.
```bash
wannagit add <path>
```

---

#### catFile
Prints the raw uncompressed object data to stdout
```bash
wannagit catFile <type> <object_hash>
```
type = blob, commit, tag

---

#### checkIgnore
check path(s) against ignore rules and prints them to stdout.
```bash
wannagit checkIgnore <path>
```

---

#### checkout
Checkout a commit inside of an empty directory.
```bash
wannagit checkout <empty_directory>
```

---

#### commit
Record changes to the repository
```bash
wannagit commit [-m <message>] 
```
flags:
-m, --message string  gives the message to describe the commit

---

#### hashObject
Compute object hash and optionally create an object from a file
```bash
wannagit hashObject [-w] [-t TYPE] <file>
```

flags: 
-t, --type string     gives the type of object (default "blob") 
-w, --write bool      writes the object to wannagit directory

---

#### log
Review logging of commit data and its metadata.
creates a log.dot file which can be used to see the commit graph.
use dot to create the pdf with the graph or use an online viewer.
```bash
dot -O -Tpdf log.dot
```
```bash
wannagit log <commit_hash>
```

---

#### lsFiles
Show information about files in the index and the working tree
```bash
wannagit lsFiles [-v|--verbose]
```

flags:
-v, --verbose bool    list out all the info about the files in the staging area

---

#### lsTree
List the contents of a tree object
```bash
wannagit lsTree [-r] <object_hash>
```

flags:
-r, --recursive bool     recursively prints the blobs instead of the tree objects

---

#### revParse
Parse revision (or other objects) identifiers 
```bash
wannagit revParse [--type <type>] <reference> 
```

flags:
-t, --type string     give the type of object (default "blob")

---

#### rm
Remove files from the working tree and from the index
```bash
wannagit rm <paths>
```

---

#### showRef
Shows all the references to commit files in the repository
```
wannagit showRef
```

---

#### status
Show the working tree status
```bash
wannagit status
```
---

#### tag
Add a reference in refs/tags/
```bash
wannagit tag <name> <object> [-a]
```

flags:
-a, --storeTrue bool     create a new tag object pointing at HEAD or OBJECT

---
