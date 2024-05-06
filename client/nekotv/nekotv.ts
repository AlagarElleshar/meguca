import {connSM, connState, message, sendBinary} from "../connection";
import {escape} from "../util"
import {Player} from "./player";

export let player: Player;

let playlistDiv: HTMLDivElement;
let playlistOl: HTMLOListElement;
let playerDiv: HTMLDivElement;
let playlistStatus: HTMLElement;
export let vidEl: HTMLVideoElement;
let watchStatus: HTMLElement;
let currentSource: string;
let watchDiv: HTMLElement;
let playerTimeInterval: number | null = null;
let nekoTV = document.getElementById("banner-nekotv");
let isOpen : boolean;
let isPlaylistVisible = false;
let subscribeMessage = new Uint8Array([1,message.nekoTV]).buffer
let unsubMessage = new Uint8Array([0,message.nekoTV]).buffer
let isMuted : boolean;

export function initNekoTV() {
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
    playerDiv.addEventListener("click",()=>{
        let is_coarse = matchMedia('(pointer:coarse)').matches
        if(is_coarse){
            return
        }
        if (playlistDiv.style.display) {
            playlistDiv.style.display = ''
        } else {
            playlistDiv.style.display = 'block'
        }
    })
    let lastVal = localStorage.getItem('neko-tv')
    if (lastVal) {
        isOpen = lastVal === 't';
    } else {
        isOpen = true;
    }
    updateNekoTVIcon()
    connSM.on(connState.synced,subscribeToWatchFeed)
    nekoTV.addEventListener("click", () => {
        isOpen = !isOpen;
        localStorage.setItem('neko-tv', isOpen ? 't' : 'f');
        updateNekoTVIcon()
        togglePlayer()
    });

    let watchCloseButton = document.getElementById('watch-close-button');
    let watchMuteButton = document.getElementById('watch-mute-button');
    watchCloseButton.addEventListener('click',()=>{
        isOpen = false;
        localStorage.setItem('neko-tv', 'f');
        updateNekoTVIcon()
        togglePlayer()
    })
    lastVal = localStorage.getItem('neko-tv-mute')
    if (lastVal) {
        isMuted = lastVal === 't';
    } else {
        isMuted = false;
    }
    if(isMuted) {
        watchMuteButton.innerText = '􀊢'
        watchMuteButton.title = 'Unmute'
    }
    else {
        watchMuteButton.innerText = '􀊦'
        watchMuteButton.title = 'Mute'
    }
    watchMuteButton.addEventListener('click',()=> {
        isMuted = !isMuted;
        localStorage.setItem('neko-tv-mute', isMuted ? 't' : 'f');
        player.setMuted(isMuted)
        if(isMuted) {
            watchMuteButton.innerText = '􀊢'
            watchMuteButton.title = 'Mute'
        }
        else {
            watchMuteButton.innerText = '􀊦'
            watchMuteButton.title = 'Unmute'
        }
    })
    player = new Player()

}

export function isNekoTVOpen() {
    return isOpen;
}

export function isNekoTVMuted() {
    return isMuted;
}

function updateNekoTVIcon(){
    if (isOpen) {
        nekoTV.innerText = '􀵨';
        nekoTV.title = 'NekoTV: Enabled'
    } else {
        nekoTV.innerText = '􁋞';
        nekoTV.title = 'NekoTV: Disabled'
    }

}
export function showWatchPanel() {
    watchDiv.style.display = 'flex';
    watchDiv.classList.remove('hide-watch-panel');
}

export function hideWatchPanel() {
    watchDiv.classList.add('hide-watch-panel');
    watchDiv.style.display = 'none';
}
export function showPlaylist() {
    playlistDiv.style.display = 'block';
}

export function hidePlaylist() {
    playlistDiv.style.display = 'none';
    stopPlayerTimeInterval();
}

export function toggleNekoTV(){
    isOpen = !isOpen;
    localStorage.setItem('neko-tv', isOpen ? 't' : 'f');
    updateNekoTVIcon()
    togglePlayer()
}

export function togglePlaylist() {
    isPlaylistVisible = !isPlaylistVisible;
    if (isPlaylistVisible) {
        showPlaylist();
    } else {
        hidePlaylist();
    }
}


export function updatePlayerTime() {
    if (!playlistOl || !playlistOl.firstElementChild ) {
        console.log('Skipping updatePlayerTime');
        return;
    }

    let playerTime = player.getTime();

    if (playerTime === undefined || playerTime === null) {
        console.error('Player time undefined');
        return;
    }

    playlistOl.children[player.getItemPos()].querySelector('.watch-player-time')!.innerHTML = `${secondsToTimeExact(playerTime)} / `;
}

function stopPlayerTimeInterval() {
    if (playerTimeInterval) {
        clearInterval(playerTimeInterval);
        playerTimeInterval = null;
    }
}

export function secondsToTimeExact(totalSeconds: number): string {
    totalSeconds = Math.floor(totalSeconds);

    const hours = Math.floor(totalSeconds / 3600);
    const minutes = Math.floor((totalSeconds - hours * 3600) / 60);
    const seconds = Math.round(totalSeconds - hours * 3600 - minutes * 60);

    let formattedTime: string;

    if (hours) {
        formattedTime = `${hours}:${padWithZero(minutes)}:${padWithZero(seconds)}`;
    } else if (minutes) {
        formattedTime = `${minutes}:${padWithZero(seconds)}`;
    } else {
        formattedTime = `0:${padWithZero(seconds)}`;
    }

    return formattedTime;
}

function padWithZero(value: number): string {
    return value < 10 ? `0${value}` : value.toString();
}
export function updatePlaylist() {
    if (player.isListEmpty()) {
        removePlayer()
        return;
    }
    if (!playerTimeInterval) {
        updatePlayerTime();
        playerTimeInterval = setInterval(updatePlayerTime, 1000);
    }

    // updatePlaylistStatus();
    showWatchPanel()

    const playlistItems: HTMLLIElement[] = [];

    const currentItemPos = player.getItemPos();

    for (let i = 0; i < player.videoList.items.length; i++) {
        const video = player.videoList.items[i];
        const li = document.createElement('li');
        li.classList.add('watch-playlist-entry');

        if (i === currentItemPos) {
            li.classList.add('selected');
        }

        let videoTerm = '';
        if (video.url && !video.url.startsWith('https')) {
            videoTerm = escape(video.url);
        }

        const videoTitle = escape(video.title);
        let durationString: string = null;
        if (video.duration == Number.POSITIVE_INFINITY) {
            durationString = '∞';
        }
        else {
            durationString = secondsToTimeExact(video.duration)
        }

        li.innerHTML = `
  <span class="watch-video-term">${videoTerm}</span>
  <a class="watch-video-title" target="_blank" href="${video.url}" title="${escape(video.title)}">
    ${videoTitle}
  </a>
  <span class="watch-video-time">
    <span class="watch-player-time"></span>
    <span class="watch-player-dur">${durationString}</span>
  </span>
`;

        playlistItems.push(li);
    }
    playlistOl.replaceChildren(...playlistItems);

    if (!isOpen) {
        isOpen = true;
    }
}

export function togglePlayer() {
    if (isOpen) {
        subscribeToWatchFeed();
    }
    else {
        unsubscribeFromWatchFeed();
    }
}

export function unsubscribeFromWatchFeed() {
    removePlayer();
    sendBinary(unsubMessage)
}

export function subscribeToWatchFeed() {
    if (isOpen) {
        sendBinary(subscribeMessage)
    }
}

export function removePlayer() {
    player.removeVideo();
    hideWatchPanel();
}