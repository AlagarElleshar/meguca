import {message, sendBinary} from "../connection";
import {ConnectedEvent, WebSocketMessage} from "../typings/messages";
import {Player} from "./player";

let player: Player;

export function initNekoTV(){
    let nekoTV = document.getElementById("banner-nekotv")
    if (!nekoTV){
        return;
    }
    nekoTV.addEventListener("click", () => {
        sendBinary(new Uint8Array([message.nekoTV]));
    })
}

function handleConnectedEvent(connectedEvent: ConnectedEvent) {
    let vl = player.videoList
    vl.setItems(connectedEvent.videoList)
    vl.setPos(connectedEvent.itemPos)
    player.onSync(connectedEvent.getTime.time)
}


export function handleMessage(message:WebSocketMessage){
    if (message.connectedEvent) {
        handleConnectedEvent(message.connectedEvent);
    }
    else if (message.addVideoEvent) {
        player.videoList.addItem(message.addVideoEvent.item, message.addVideoEvent.atEnd)
    }
    else if (message.removeVideoEvent) {
        player.removeItem(message.removeVideoEvent.url)
    }
    else if (message.skipVideoEvent) {
        player.skipVideo(message.skipVideoEvent.url);
    }
    else if (message.pauseEvent) {
        handlePauseEvent(message.pauseEvent);
    }
    else if (message.playEvent) {
        handlePlayEvent(message.playEvent);
    }
    else if (message.getTimeEvent) {
        handleGetTimeEvent(message.getTimeEvent);
    }
    else if (message.setTimeEvent) {
        handleSetTimeEvent(message.setTimeEvent);
    }
    else if (message.setRateEvent) {
        handleSetRateEvent(message.setRateEvent);
    }
    else if (message.rewindEvent) {
        handleRewindEvent(message.rewindEvent);
    }
    else if (message.playItemEvent) {
        handlePlayItemEvent(message.playItemEvent);
    }
    else if (message.setNextItemEvent) {
        handleSetNextItemEvent(message.setNextItemEvent);
    }
    else if (message.updatePlaylistEvent) {
        handleUpdatePlaylistEvent(message.updatePlaylistEvent);
    }
    else if (message.togglePlaylistLockEvent) {
        handleTogglePlaylistLockEvent(message.togglePlaylistLockEvent);
    }
    else if (message.dumpEvent) {
        handleDumpEvent(message.dumpEvent);
    }
    else if (message.clearPlaylistEvent) {
        handleClearPlaylistEvent(message.clearPlaylistEvent);
    }
    else {
        console.error('Invalid WebSocketMessage received');
    }
}