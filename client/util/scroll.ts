// Various page scrolling aids

import { page } from "../state"
import { trigger } from "./hooks"
import { lightenThread } from "../posts";

const banner = document.getElementById("banner")

let scrolled = false
let locked = false;

// Indicates if the page is scrolled to its bottom
export let atBottom: boolean

// Scroll to target anchor element, if any
export function scrollToAnchor() {
	if (!location.hash) {
		if (!page.thread) {
			scrollToTop()
		}
		return
	}
	const el = document.querySelector(location.hash) as HTMLElement
	if (!el) {
		return scrollToTop()
	}
	scrollToElement(el)
	checkBottom()
}

function isTheaterModeActive() {
	return document.body.classList.contains("nekotv-theater")
}

function getTheaterModeRightDiv() {
	return document.getElementById("right-content")
}

// Scroll to particular element and compensate for the banner height
export function scrollToElement(el: HTMLElement) {
	if(isTheaterModeActive()){
		let rightDiv = getTheaterModeRightDiv()
		rightDiv.scrollTo(0, el.offsetTop - rightDiv.offsetTop - 5)
	}
	else {
		window.scrollTo(0, el.offsetTop - banner.offsetHeight - 5)
	}
}

function scrollToTop() {
	if(isTheaterModeActive()){
		let rightDiv = getTheaterModeRightDiv()
		rightDiv.scrollTo(0, 0)
	}
	else {
		window.scrollTo(0, 0)
	}
	checkBottom()
}

// Scroll to the bottom of the thread
export function scrollToBottom() {
	if(isTheaterModeActive()){
		let rightDiv = getTheaterModeRightDiv()
		rightDiv.scrollTo(0, rightDiv.scrollHeight)
	}
	else {
		window.scrollTo(0, document.documentElement.scrollHeight)
	}
	atBottom = true
}

// Check, if at the bottom of the thread and render the locking indicator
export function checkBottom() {
	if (!page.thread) {
		atBottom = false
		return
	}
	const previous = atBottom;
	atBottom = isAtBottom()
	const lock = document.getElementById("lock")
	if (lock) {
		lock.style.visibility = atBottom ? "visible" : "hidden"
	}
	if (!previous && atBottom) {
		lightenThread();
	}
}

// Return, if scrolled to bottom of page
export function isAtBottom(): boolean {
	if(isTheaterModeActive()){
		let rightDiv = getTheaterModeRightDiv()
		let {scrollTop, scrollHeight, offsetHeight} = rightDiv
		console.log(scrollTop, scrollHeight, offsetHeight)
		return Math.abs(rightDiv.scrollTop - (rightDiv.scrollHeight - rightDiv.offsetHeight)) < 2
	}
	return window.innerHeight
		+ window.scrollY
		- document.documentElement.offsetHeight
		> -1
}
// @ts-ignore
window.isAtBottom = isAtBottom

// If we are at the bottom, lock
document.addEventListener("scroll", () => {
	scrolled = !isAtBottom()
	locked = !scrolled;
	checkBottom();
}, { passive: true })

export function addTheaterModeScrollListener(){
	let rightDiv = getTheaterModeRightDiv()
	rightDiv.addEventListener("scroll", () => {
		scrolled = !isAtBottom()
		locked = !scrolled;
		checkBottom();
	}, { passive: true })
}

// Use a MutationObserver to jump to the bottom of the page when a new
// post is made, we are locked to the bottom or the user set the alwaysLock option
let threadContainer = document.getElementById("thread-container")
if (threadContainer !== null) {
	let threadObserver = new MutationObserver((mut) => {
		if (locked || (trigger("getOptions").alwaysLock)) {
			scrollToBottom()
		}
	})
	threadObserver.observe(threadContainer, {
		childList: true,
		subtree: true,
	})
}

// Unlock from bottom, when the tab is hidden
document.addEventListener("visibilitychange", () => {
	if (document.hidden) {
		locked = false
	}
})

window.addEventListener("hashchange", scrollToAnchor, {
	passive: true,
})
