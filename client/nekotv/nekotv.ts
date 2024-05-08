import {connSM, connState, message, sendBinary} from "../connection";
import {Player} from "./player";
import {getTheaterMode, setTheaterMode} from "./theaterMode";
import {togglePlaylist, updatePlaylist} from "./playlist";



export let player: Player;
export let watchVideoDiv = document.getElementById('watch-video') as HTMLVideoElement;

let playerDiv = document.getElementById('watch-player') as HTMLDivElement;
let playlistStatus = document.getElementById('watch-playlist-status')!;
let watchStatus = document.getElementById('status-watch')!;
let watchDiv = document.getElementById("watch-panel");
let nekoTVBannerIcon : HTMLElement;

let subscribeMessage = new Uint8Array([1,message.nekoTV]).buffer
let unsubMessage = new Uint8Array([0,message.nekoTV]).buffer

let isMuted : boolean;
let isNekoTVEnabled : boolean;
let panelVisible : boolean = false;
export let watchPlaylistButton: HTMLElement;
export let watchMuteButton: HTMLElement;

export function initNekoTV() {

    nekoTVBannerIcon = document.getElementById("banner-nekotv");
    playerDiv.addEventListener("click",()=>{
        let is_coarse = matchMedia('(pointer:coarse)').matches
        if(is_coarse){
            return
        }
        togglePlaylist()
    })
    let lastVal = localStorage.getItem('neko-tv')
    if (lastVal) {
       isNekoTVEnabled= lastVal === 't';
    } else {
        isNekoTVEnabled = true;
    }
    updateNekoTVIcon()
    connSM.on(connState.synced,subscribeToWatchFeed)
    nekoTVBannerIcon.addEventListener("click", () => {
        setNekoTVEnabled(!isNekoTVEnabled)
    });

    let watchCloseButton = document.getElementById('watch-close-button');
    watchMuteButton = document.getElementById('watch-mute-button');
    watchPlaylistButton = document.getElementById('watch-playlist-button');
    let watchTheaterButton = document.getElementById('watch-theater-button');
    watchCloseButton.addEventListener('click',()=>{
        setNekoTVEnabled(false)
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
    watchPlaylistButton.addEventListener('click',()=>{
        togglePlaylist()
    });
    watchTheaterButton.addEventListener('click',()=>{
        setTheaterMode(!getTheaterMode())
    })
    player = new Player()
}

export function updateNekoTVPanel(){
    if (player.isListEmpty() || !isNekoTVEnabled) {
        setTheaterMode(false)
        setPanelVisible(false)
        player.removeVideo()
    }
    else{
        setPanelVisible(true)
        updatePlaylist()
    }
}

export function isNekoTVMuted() {
    return isMuted;
}

function updateNekoTVIcon(){
    if (isNekoTVEnabled) {
        nekoTVBannerIcon.innerText = '􀵨';
        nekoTVBannerIcon.title = 'NekoTV: Enabled'
    } else {
        nekoTVBannerIcon.innerText = '􁋞';
        nekoTVBannerIcon.title = 'NekoTV: Disabled'
    }
}

function setPanelVisible(visible: boolean){
    if(panelVisible == visible){
        return
    }
    panelVisible = visible;
    if(visible){
        watchDiv.style.display = 'flex';
        watchDiv.classList.remove('hide-watch-panel');
    }
    else{
        watchDiv.classList.add('hide-watch-panel');
        watchDiv.style.display = 'none';
    }
}

export function toggleNekoTV() {
    setNekoTVEnabled(!isNekoTVEnabled)
}
function setNekoTVEnabled(value: boolean){
    if(isNekoTVEnabled == value){
        return
    }
    isNekoTVEnabled = value;
    updateNekoTVIcon()
    localStorage.setItem('neko-tv', isNekoTVEnabled ? 't' : 'f');
    updateNekoTVPanel()
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

export function unsubscribeFromWatchFeed() {
    sendBinary(unsubMessage)
}

export function subscribeToWatchFeed() {
    if (isNekoTVEnabled) {
        sendBinary(subscribeMessage)
    }
}