import {
    AddVideoEvent,
    ClearPlaylistEvent, ConnectedEvent,
    DumpEvent, GetTimeEvent, PauseEvent, PlayEvent,
    PlayItemEvent, RemoveVideoEvent,
    RewindEvent,
    SetNextItemEvent,
    SetRateEvent,
    SetTimeEvent, SkipVideoEvent, TogglePlaylistLockEvent,
    UpdatePlaylistEvent, WebSocketMessage
} from "../typings/messages";

import {player, updateNekoTVPanel} from "./nekotv";

export function handleMessage(message: WebSocketMessage) {
    switch (message.messageType.oneofKind) {
        case "connectedEvent":
            handleConnectedEvent(message.messageType.connectedEvent);
            break;
        case "addVideoEvent":
            handleAddVideoEvent(message.messageType.addVideoEvent);
            break;
        case "removeVideoEvent":
            handleRemoveVideoEvent(message.messageType.removeVideoEvent);
            break;
        case "skipVideoEvent":
            handleSkipVideoEvent(message.messageType.skipVideoEvent);
            break;
        case "pauseEvent":
            handlePauseEvent(message.messageType.pauseEvent);
            break;
        case "playEvent":
            handlePlayEvent(message.messageType.playEvent);
            break;
        case "getTimeEvent":
            handleGetTimeEvent(message.messageType.getTimeEvent);
            break;
        case "setTimeEvent":
            handleSetTimeEvent(message.messageType.setTimeEvent);
            break;
        case "setRateEvent":
            handleSetRateEvent(message.messageType.setRateEvent);
            break;
        case "rewindEvent":
            handleRewindEvent(message.messageType.rewindEvent);
            break;
        case "playItemEvent":
            handlePlayItemEvent(message.messageType.playItemEvent);
            break;
        case "setNextItemEvent":
            handleSetNextItemEvent(message.messageType.setNextItemEvent);
            break;
        case "updatePlaylistEvent":
            handleUpdatePlaylistEvent(message.messageType.updatePlaylistEvent);
            break;
        case "togglePlaylistLockEvent":
            handleTogglePlaylistLockEvent(message.messageType.togglePlaylistLockEvent);
            break;
        case "dumpEvent":
            handleDumpEvent(message.messageType.dumpEvent);
            break;
        case "clearPlaylistEvent":
            handleClearPlaylistEvent(message.messageType.clearPlaylistEvent);
            break;
        default:
            console.error("Invalid WebSocketMessage received");
    }
}

function handleSetTimeEvent(setTimeEvent: SetTimeEvent) {
    const synchThreshold = 1600;
    const newTime = setTimeEvent.time;
    const time = player.getTime();
    if (Math.abs(time - newTime) < synchThreshold) {
        return;
    }
    player.setTime(newTime);
}

function handleSetRateEvent(setRateEvent: SetRateEvent) {
    player.setPlaybackRate(setRateEvent.rate);
}

function handleRewindEvent(rewindEvent: RewindEvent) {
    player.setTime(rewindEvent.time + 0.5);
}

function handlePlayItemEvent(playItemEvent: PlayItemEvent) {
    player.setVideo(playItemEvent.pos);
    updateNekoTVPanel()
}

function handleSetNextItemEvent(setNextItemEvent: SetNextItemEvent) {
    player.setNextItem(setNextItemEvent.pos);
    updateNekoTVPanel()
}

function handleUpdatePlaylistEvent(updatePlaylistEvent: UpdatePlaylistEvent) {
    player.setItems(updatePlaylistEvent.videoList.items);
    updateNekoTVPanel()
}

function handleTogglePlaylistLockEvent(togglePlaylistLockEvent: TogglePlaylistLockEvent) {
    // player.setPlaylistLock(togglePlaylistLockEvent.isOpen);
}

function handleDumpEvent(dumpEvent: DumpEvent) {
    // Implement the logic for handling the dump event if needed
}

function handleClearPlaylistEvent(clearPlaylistEvent: ClearPlaylistEvent) {
    player.clearItems();
    updateNekoTVPanel()
}

function handleConnectedEvent(connectedEvent: ConnectedEvent) {
    player.setItems(connectedEvent.videoList,connectedEvent.itemPos)
    updateNekoTVPanel()
    handleSetTimeEvent(connectedEvent.getTime)
}

function handleAddVideoEvent(addVideoEvent: AddVideoEvent) {
    player.videoList.addItem(addVideoEvent.item, addVideoEvent.atEnd);
    if (player.itemsLength() == 1) player.setVideo(0);
    updateNekoTVPanel()
}

function handleRemoveVideoEvent(removeVideoEvent: RemoveVideoEvent) {
    player.removeItem(removeVideoEvent.url);
    updateNekoTVPanel()
}

function handleSkipVideoEvent(skipVideoEvent: SkipVideoEvent) {
    player.skipItem(skipVideoEvent.url);
    updateNekoTVPanel()
}

function handlePauseEvent(pauseEvent: PauseEvent) {
    player.pause();
    player.setTime(pauseEvent.time);
}

function handlePlayEvent(playEvent: PlayEvent) {
    // player.setPauseIndicator(true);
    // const synchThreshold = player.settings.synchThreshold;
    const newTime = playEvent.time;
    const time = player.getTime();
    if (Math.abs(time - newTime) >= 1600) {
        player.setTime(newTime);
    }
    player.play();
}

function handleGetTimeEvent(getTimeEvent: GetTimeEvent) {
    const paused = getTimeEvent.paused ?? false;
    const rate = getTimeEvent.rate ?? 1;

    if (player.getPlaybackRate() !== rate) {
        console.log('Updating playback rate to:', rate);
        player.setPlaybackRate(rate);
    }

    const synchThreshold = 1.6;
    const newTime = getTimeEvent.time;
    const time = player.getTime();

    if (!player.isVideoLoaded()) {
        console.log('Video not loaded');
        // player.forceSyncNextTick = false;
    }
    if (player.getDuration() <= time + synchThreshold) {
        console.log('Video near end, skipping synchronization');
        return;
    }
    if (!paused) {
        player.play();
    } else {
        player.pause();
    }
    // player.setPauseIndicator(!paused);
    if (Math.abs(time - newTime) < synchThreshold) {
        return;
    }
    if (!paused) {
        console.log('Synchronizing time to:', newTime + 0.5);
        player.setTime(newTime + 0.5);
    } else {
        console.log('Synchronizing time to:', newTime);
        player.setTime(newTime);
    }
}
