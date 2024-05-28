import {escape} from "../util";
import {
    player,
    secondsToTimeExact,
} from "./nekotv";

const playlistOl = document.getElementById('watch-playlist-entries') as HTMLOListElement;
const playlistDiv = document.getElementById('watch-playlist') as HTMLDivElement;
let playerTimeInterval: number | null = null;
let isPlaylistVisible = false;


export function updatePlaylist() {

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
        let moreClasses = ""
        if (video.duration == Number.POSITIVE_INFINITY) {
            durationString = '∞';
            moreClasses = " infinite"
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
    <span class="watch-player-dur${moreClasses}">${durationString}</span>
  </span>
`;

        playlistItems.push(li);
    }
    playlistOl.replaceChildren(...playlistItems);
}
export function startPlayerTimeInterval() {
    if (!playerTimeInterval) {
        updatePlayerTime();
        playerTimeInterval = setInterval(updatePlayerTime, 1000);
    }
}
export function stopPlayerTimeInterval() {
    if (playerTimeInterval) {
        clearInterval(playerTimeInterval);
        playerTimeInterval = null;
    }
}

export function showPlaylist() {
    playlistDiv.style.display = 'block';
}

export function hidePlaylist() {
    playlistDiv.style.display = '';
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

    const playerTime = player.getTime();

    if (playerTime === undefined || playerTime === null) {
        console.error('Player time undefined');
        return;
    }

    playlistOl.children[player.getItemPos()].querySelector('.watch-player-time')!.innerHTML = `${secondsToTimeExact(playerTime)} / `;
}
