#!/bin/bash

carton install --deployment
carton exec plackup -s Starlet --max-workers 10 -p 8000 app.psgi
