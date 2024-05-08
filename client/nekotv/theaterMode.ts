import {addTheaterModeScrollListener, isAtBottom, scrollToBottom} from "../util";
import { player} from "./nekotv";

let isTheaterMode = false;
let rightDiv: HTMLElement = null;

export function setTheaterMode(value: boolean) {
    if (isTheaterMode === value) {
        return;
    }
    isTheaterMode = value;
    if (value) {
        activateTheaterMode()
    } else {
        deactivateTheaterMode()
    }
}

export function getTheaterMode() {
    return isTheaterMode;
}

function activateTheaterMode() {
    const articles = document.getElementsByTagName('article');
    const atBottom = isAtBottom()

    let articleShown = null;
    for (let i = articles.length - 1; i >= 0; i--) {
        const article = articles[i];
        const rect = article.getBoundingClientRect();

        if (
            rect.top >= 0 &&
            rect.left >= 0 &&
            rect.bottom <= (window.innerHeight || document.documentElement.clientHeight) &&
            rect.right <= (window.innerWidth || document.documentElement.clientWidth)
        ) {
            articleShown = article;
            break
        }
    }

    const bodyChildren = document.body.children;
    rightDiv = document.createElement('div');
    rightDiv.id = 'right-content';
    for (let i = 0; i < bodyChildren.length; i++) {
        const child = bodyChildren[i];
        rightDiv.appendChild(child);
        i--;
    }

    document.body.appendChild(rightDiv);
    addTheaterModeScrollListener()
    const videoElement = document.getElementById('watch-panel');
    document.body.insertBefore(videoElement, document.body.firstChild);
    document.body.classList.add("nekotv-theater")
    if (atBottom) {
        rightDiv.scrollTo(0, rightDiv.scrollHeight)
    } else {
        articleShown.scrollIntoView(
            {
                behavior: "instant",
                block: "end",
            }
        )
    }
    rightDiv.scrollLeft = 0;
    player.reload()
}

function deactivateTheaterMode() {
    const watchPanel = document.getElementById('watch-panel');
    const articles = document.getElementsByTagName('article');
    const atBottom = isAtBottom()
    let articleShown = null;
    for (let i = articles.length - 1; i >= 0; i--) {
        const article = articles[i];
        const rect = article.getBoundingClientRect();

        if (
            rect.top >= 0 &&
            rect.left >= 0 &&
            rect.bottom <= (window.innerHeight || document.documentElement.clientHeight) &&
            rect.right <= (window.innerWidth || document.documentElement.clientWidth)
        ) {
            articleShown = article;
            break
        }
    }

    document.getElementById("watcher").after(watchPanel);

    while (rightDiv.firstChild) {
        document.body.appendChild(rightDiv.firstChild);
    }
    rightDiv.remove()
    rightDiv = null;

    document.body.classList.remove("nekotv-theater");
    articleShown.scrollIntoView(
        {
            behavior: "instant",
            block: "end",
        }
    )
    player.reload()
    if (atBottom) {
        scrollToBottom()
    } else {
        articleShown.scrollIntoView(
            {
                behavior: "instant",
                block: "end",
                inline: "start"
            }
        )
    }
}
