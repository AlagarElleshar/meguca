import { message, sendBinary } from "../connection";
import {
    ConnectedEvent,
    WebSocketMessage,
    AddVideoEvent,
    RemoveVideoEvent,
    SkipVideoEvent,
    PauseEvent,
    PlayEvent,
    GetTimeEvent,
    SetTimeEvent,
    SetRateEvent,
    RewindEvent,
    PlayItemEvent,
    SetNextItemEvent,
    UpdatePlaylistEvent,
    TogglePlaylistLockEvent,
    DumpEvent,
    ClearPlaylistEvent,
} from "../typings/messages";
import { Player } from "./player";

let player: Player;

export let playlistDiv: HTMLDivElement;
export let playlistOl: HTMLOListElement;
export let playerDiv: HTMLDivElement;
export let playlistStatus: HTMLElement;
export let vidEl: HTMLVideoElement;
export let watchStatus: HTMLElement;
export let currentSource: string;
export let watchDiv: HTMLElement;
let isOpen = false;

export function initNekoTV() {
    let nekoTV = document.getElementById("banner-nekotv");
    if (!nekoTV) {
        return;
    }
    playlistDiv = document.getElementById('watch-playlist') as HTMLDivElement;
    playlistOl = document.getElementById('watch-playlist-entries') as HTMLOListElement;
    playerDiv = document.getElementById('watch-player') as HTMLDivElement;
    playlistStatus = document.getElementById('watch-playlist-status')!;
    vidEl = document.getElementById('watch-video') as HTMLVideoElement;
    watchStatus = document.getElementById('status-watch')!;
    watchDiv = document.getElementById("watch-panel");
    nekoTV.addEventListener("click", () => {
        isOpen = !isOpen;
        if (isOpen) {
            sendBinary(new Uint8Array([message.nekoTV]));
            showWatchPanel();
        } else {
            hideWatchPanel();
        }
    });
    player = new Player()

}
// export function initWatch() {
//     if (watchEnabled()) {
//         playlistDiv = document.getElementById('watch-playlist') as HTMLDivElement;
//         playlistOl = document.getElementById('watch-playlist-entries') as HTMLOListElement;
//         playerDiv = document.getElementById('watch-player') as HTMLDivElement;
//         playlistStatus = document.getElementById('watch-playlist-status')!;
//         vidEl = document.getElementById('watch-video') as HTMLVideoElement;
//         watchStatus = document.getElementById('status-watch')!;
//
//         embedFunctions = {
//             [mediaSource.YouTube]: embedYouTube,
//         };
//
//         handlers[message.watchData] = watchMessageHandler;
//
//         watchStatus.addEventListener('mouseover', () => {
//             if (fullTitle) {
//                 watchStatus.textContent = fullTitle;
//                 showPlaylist();
//             }
//         }, {passive: true});
//
//         watchStatus.addEventListener('mouseout', () => {
//             if (!isPlaylistVisible) {
//                 watchStatus.textContent = truncatedTitle;
//                 hidePlaylist();
//             }
//         }, {passive: true});
//
//         watchStatus.addEventListener('click', (e) => {
//             isPlaylistVisible = !isPlaylistVisible;
//             if (isPlaylistVisible) {
//                 showPlaylist();
//             } else {
//                 hidePlaylist();
//             }
//             e.preventDefault();
//         });
//
//         watchDiv.addEventListener('mouseover', () => {
//             bodyClassList.add('player-hovered');
//             if (!localStorage.watchTipShown) {
//                 tempNotify(
//                     'Live-Synced Video Player Tips',
//                     'Press &quot;<code class="tip-highlight">Alt+W</code>&quot; (<code class="tip-highlight">W</code> = &quot;watch&quot;) to hide/show the video player.<br><br>Press &quot;<code class="tip-highlight">Alt+M</code>&quot; (<code class="tip-highlight">M</code> = &quot;mute&quot;) to mute/unmute audio for the video player (and any other audio on the site). Click the &quot;Settings&quot; gear icon in the top-right to adjust the volume.<br><br>Hover over the video title in the top-right to show the current playlist. Click the video title to toggle persistent display of the playlist, or press &quot;<code class="tip-highlight">Alt+L</code>&quot; (<code class="tip-highlight">L</code> = play&quot;list&quot;).',
//                     'watch-tips',
//                     90
//                 );
//                 localStorage.watchTipShown = '1';
//             }
//         }, {passive: true});
//
//         watchDiv.addEventListener('mouseout', () => {
//             bodyClassList.remove('player-hovered');
//         }, {passive: true});
//
//         // @ts-ignore
//         socket.connSM.on(3, subscribeToWatchFeed);
//     }
// }
export function showWatchPanel() {

    watchDiv.style.display = 'block';
    watchDiv.classList.remove('hide-watch-panel');
}

export function hideWatchPanel() {
    watchDiv.classList.add('hide-watch-panel');
}

function handleConnectedEvent(connectedEvent: ConnectedEvent) {
    player.setItems(connectedEvent.videoList,connectedEvent.itemPos)
    handleSetTimeEvent(connectedEvent.getTime)
}

function handleAddVideoEvent(addVideoEvent: AddVideoEvent) {
    player.videoList.addItem(addVideoEvent.item, addVideoEvent.atEnd);
    if (player.itemsLength() === 1) {
        player.setVideo(0);
    }
}

function handleRemoveVideoEvent(removeVideoEvent: RemoveVideoEvent) {
    player.removeItem(removeVideoEvent.url);
    if (player.isListEmpty()) {
        player.pause();
    }
}

function handleSkipVideoEvent(skipVideoEvent: SkipVideoEvent) {
    player.skipItem(skipVideoEvent.url);
    if (player.isListEmpty()) player.pause();
}

function handlePauseEvent(pauseEvent: PauseEvent) {
    // player.setPauseIndicator(false);
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
        player.setPlaybackRate(rate);
    }

    const synchThreshold = 1600;
    const newTime = getTimeEvent.time;
    const time = player.getTime();

    if (!player.isVideoLoaded()) {
        // player.forceSyncNextTick = false;
    }
    if (player.getDuration() <= time + synchThreshold) {
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
        player.setTime(newTime + 0.5);
    } else {
        player.setTime(newTime);
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
}

function handleSetNextItemEvent(setNextItemEvent: SetNextItemEvent) {
    player.setNextItem(setNextItemEvent.pos);
}

function handleUpdatePlaylistEvent(updatePlaylistEvent: UpdatePlaylistEvent) {
    player.setItems(updatePlaylistEvent.videoList.items);
}

function handleTogglePlaylistLockEvent(togglePlaylistLockEvent: TogglePlaylistLockEvent) {
    // player.setPlaylistLock(togglePlaylistLockEvent.isOpen);
}

function handleDumpEvent(dumpEvent: DumpEvent) {
    // Implement the logic for handling the dump event if needed
}

function handleClearPlaylistEvent(clearPlaylistEvent: ClearPlaylistEvent) {
    player.clearItems();
    if (player.isListEmpty()) {
        player.pause();
    }
}

export function handleMessage(message: WebSocketMessage) {
    if (message.connectedEvent) {
        handleConnectedEvent(message.connectedEvent);
    } else if (message.addVideoEvent) {
        handleAddVideoEvent(message.addVideoEvent);
    } else if (message.removeVideoEvent) {
        handleRemoveVideoEvent(message.removeVideoEvent);
    } else if (message.skipVideoEvent) {
        handleSkipVideoEvent(message.skipVideoEvent);
    } else if (message.pauseEvent) {
        handlePauseEvent(message.pauseEvent);
    } else if (message.playEvent) {
        handlePlayEvent(message.playEvent);
    } else if (message.getTimeEvent) {
        handleGetTimeEvent(message.getTimeEvent);
    } else if (message.setTimeEvent) {
        handleSetTimeEvent(message.setTimeEvent);
    } else if (message.setRateEvent) {
        handleSetRateEvent(message.setRateEvent);
    } else if (message.rewindEvent) {
        handleRewindEvent(message.rewindEvent);
    } else if (message.playItemEvent) {
        handlePlayItemEvent(message.playItemEvent);
    } else if (message.setNextItemEvent) {
        handleSetNextItemEvent(message.setNextItemEvent);
    } else if (message.updatePlaylistEvent) {
        handleUpdatePlaylistEvent(message.updatePlaylistEvent);
    } else if (message.togglePlaylistLockEvent) {
        handleTogglePlaylistLockEvent(message.togglePlaylistLockEvent);
    } else if (message.dumpEvent) {
        handleDumpEvent(message.dumpEvent);
    } else if (message.clearPlaylistEvent) {
        handleClearPlaylistEvent(message.clearPlaylistEvent);
    } else {
        console.error("Invalid WebSocketMessage received");
    }
}