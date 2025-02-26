#!/bin/bash

# Define the paths for the JSON and YAML files
STORAGE_JSON="storage.json"
YAML_FILE="Supfile"
TMP_FILE="Supfile.tmp"

# Extract IDs and hostnames from storage.json
echo "Extracting IDs and hostnames from $STORAGE_JSON..."
jq -r '.[] | "- \(.ip) # \(.hostname // "unknown")"' "$STORAGE_JSON" > id_list.tmp
echo "ID List:"
cat id_list.tmp

# Use awk to insert the ID list into the Supfile
echo "Updating $YAML_FILE..."
#TODO update Supfile with the new hosts


# Clean up the temporary file
rm id_list.tmp