# On demand backup

>(ODB)


----------

Have your system create a copy of your file on change.

# How it works?

You specify ODB a directory to watch. ODB will then watch all of the files within this directory. Each time there is data written to it ODB will create a copy. The backups are in the same location where you launched the program in folder `backup-appe`

# How to run it?

	odb <Directory path>


### In the background

Requires `nohup`.

	nohup odb <Directory path> &
