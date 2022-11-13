# cruddyString
![example workflow](https://github.com/charliehogger31/cruddyString/actions/workflows/codeql.yml/badge.svg)
![example workflow](https://github.com/charliehogger31/cruddyString/actions/workflows/go.yml/badge.svg)


CRUD thing were you can put strings up.

Usage: cruddyString (_inifile_) (_preload_)

# API
CruddyString (CS) hosts an HTTP server on port 8080
## CRUD
### Create
To create post data to the root directory '/'.
The input field is 'inputdata' set that to whatever you want to upload.
CS will return a number which corresponds to the index of the resource you uploaded.
### Read
To read data send a get request to the directory '/i', where i corresponds to the index of the resource.
CS will return the resource if it exists or if it does not exist it will return 400 Bad Request.
### Update
Send a PATCH request to the directory '/i' where i is the index of the resource you wish to update.
The input field is 'inputdata', same as Create.
Update will respond with the old data before the rewrite.
### Delete
Send a DELETE request tot the directory '/i' where i is the index of the resource you wish to update.
Delete will respond with the old data that was deleted.

## Ini file
An ini file can be optionally used.
```
[memory]
maxresourcesize=x
maxnumresources=y
```
> Replace x with the max resource size
> and y with the max number of resources

Put the path to the ini file in the first argument.

## Preload
You can preload resources to CS with a text file.
Supply the path to this text file in the second argument.
> Note: You must use an ini file if you wish to preload.

CS will read the file line by line adding each line as a resource.
