#!/bin/bash

#### Startup script that triggers on machine startup, put any logic here ###
set -e

WAIT_FOR_RAID_SECONDS=120 # 2 minutes max wait for RAID to be created before we copy artifacts
RAID_DEVICE="/dev/md0"
MOUNT_POINT="/home/injectived/.injectived"
ARTIFACTS_DIR="/home/injectived/artifacts"
RAID_DEVICES="/dev/nvme0n1 /dev/nvme0n2 /dev/nvme0n3 /dev/nvme0n4"

# Function to wait for /dev/md0 readiness
wait_for_raid() {
    echo "⏳ Waiting for $RAID_DEVICE to be ready..."
    for i in $(seq 1 $WAIT_FOR_RAID_SECONDS); do
        if [ -e "$RAID_DEVICE" ]; then
            echo "✅ $RAID_DEVICE exists."
            return 0
        fi
        sleep 1
    done
    echo "❌ $RAID_DEVICE was not created in time."
    return 1
}

# Wait for all required artifacts to be copied
wait_for_artifacts() {
    local max_wait=300  # 5 minutes timeout
    local required_items=(
        "data"
        "config"
        "injectived"
        "libwasmvm.x86_64.so"
    )

    echo "⏳ Waiting for all required artifacts in $ARTIFACTS_DIR..."
    for i in $(seq 1 $max_wait); do
        local all_found=true
        
        # Check each required item
        for item in "${required_items[@]}"; do
            if [ ! -e "$ARTIFACTS_DIR/$item" ]; then
                all_found=false
                echo "⏳ Waiting for $item... ($i/$max_wait seconds)"
                break
            fi
        done

        # If all items were found
        if $all_found; then
            echo "✅ All required artifacts found:"
            ls -la $ARTIFACTS_DIR
            return 0
        fi

        sleep 1
    done
    
    echo "❌ Timeout waiting for artifacts. Missing items:"
    for item in "${required_items[@]}"; do
        if [ ! -e "$ARTIFACTS_DIR/$item" ]; then
            echo "  - $item"
        fi
    done
    return 1
}

# Build RAID only if it doesn't exist
if [ ! -e "$RAID_DEVICE" ]; then
    echo "⚡ Creating RAID0 on $RAID_DEVICE..."
    sudo mdadm --create --verbose $RAID_DEVICE --level=0 --raid-devices=4 $RAID_DEVICES || exit 1

    # Wait for RAID device and proceed with setup
    if wait_for_raid; then
        # Create filesystem and mount
        echo "📝 Formatting with XFS..."
        sudo mkfs.xfs $RAID_DEVICE
        sudo mount $RAID_DEVICE $MOUNT_POINT
        sudo chown -R injectived:injectived $MOUNT_POINT

        echo "📁 Waiting for artifacts to be ready..."
        if ! wait_for_artifacts; then
            echo "❌ Failed to find artifacts"
            exit 1
        fi

        echo "📁 Copying artifacts after RAID is mounted..."
        rsync -av $ARTIFACTS_DIR/ $MOUNT_POINT/
        sudo chown -R injectived:injectived $MOUNT_POINT

        echo "🎉 RAID setup and artifact copy completed successfully!"
    else
        echo "❌ Failed to create RAID array"
        exit 1
    fi
fi