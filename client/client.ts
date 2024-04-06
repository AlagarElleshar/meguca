// Core websocket message handlers

import {handlers, message, connSM, connEvent, decoder} from './connection'
import { posts, page } from './state'
import { Post, FormModel, PostView, lightenThread } from './posts'
import { PostLink, Command, PostData, ImageData, ModerationEntry } from "./common"
import { postAdded } from "./ui"
import { decrementImageCount, incrementPostCount } from "./page"
import { getPostName } from "./options"
import { OverlayNotification } from "./ui"
import { setCookie }  from './util';
import { debug } from "./state"

// Message for splicing the contents of the current line
export type SpliceResponse = {
	id: number
	start: number
	len: number
	text: string
}

type CloseMessage = {
	id: number
	links: PostLink[] | null
	commands: Command[] | null
}

// Message for inserting images into an open post
interface ImageMessage extends ImageData {
	id: number
}

interface ModerationMessage extends ModerationEntry {
	id: number
}

interface CookieMessage {
	key: string;
	value: string;
}

// Run a function on a model, if it exists
function handle(id: number, fn: (m: Post) => void) {
	const model = posts.get(id)
	if (model) {
		fn(model)
	}
}

// Insert a post into the models and DOM
export function insertPost(data: PostData) {
	// Now playing post name override
	const postName = getPostName()
	if (postName !== undefined) {
		data.name = postName
	}

	const existing = posts.get(data.id)
	if (existing) {
		if (existing instanceof FormModel) {
			existing.onAllocation(data)
			incrementPostCount(true, !!data["image"]);
		}
		return
	}

	const model = new Post(data)
	model.op = page.thread
	model.board = page.board
	posts.add(model)
	const view = new PostView(model, null)

	if (!model.editing) {
		model.propagateLinks()
	}

	// Find last allocated post and insert after it
	const last = document
		.getElementById("thread-container")
		.lastElementChild
	if (last.id === "p0") {
		last.before(view.el)
	} else {
		last.after(view.el)
	}

	postAdded(model)
	incrementPostCount(true, !!data["image"]);
	lightenThread();
}

export default () => {
	handlers[message.invalid] = (msg: string) => {

		// TODO: More user-friendly critical error reporting

		alert(msg)
		connSM.feed(connEvent.error)
		throw msg
	}

	handlers[message.insertPost] = insertPost

	handlers[message.insertImage] = (msg: ImageMessage) =>
		handle(msg.id, m => {
			delete msg.id
			if (!m.image) {
				incrementPostCount(false, true)
			}
			m.insertImage(msg)
		})

	handlers[message.spoiler] = (id: number) =>
		handle(id, m =>
			m.spoilerImage())

	handlers[message.append] = (message: ArrayBuffer) => {
		const view = new DataView(message);
		let id = view.getFloat64(0, true);
		let append = decoder.decode(message.slice(8));
		if(debug)
			console.log(`>binary append ${id} ${append}`)
		handle(id, m =>
			m.appendString(append))
	}

	handlers[message.backspace] = (message: ArrayBuffer) => {
		const view = new DataView(message);
		let id = view.getFloat64(0, true);
		if(debug)
			console.log(`>binary backspace ${id}`)
		handle(id, m =>
			m.backspace())
	}

	handlers[message.splice] = (binaryMessage: ArrayBuffer) => {
		const view = new DataView(binaryMessage);
		let msg: SpliceResponse = {
			id: view.getFloat64(0, true),
			start: view.getUint16(8, true),
			len: view.getUint16(10, true),
			text: decoder.decode(binaryMessage.slice(12))
		}
		if(debug)
			console.log(`>binary splice ${msg}`)
		handle(msg.id, m =>
			m.splice(msg))
	}

	handlers[message.closePost] = ({ id, links, commands }: CloseMessage) =>
		handle(id, m => {
			if (links) {
				m.links = links
				m.propagateLinks()
			}
			if (commands) {
				m.commands = commands
			}
			m.closePost()
		})

	handlers[message.moderatePost] = (msg: ModerationMessage) =>
		handle(msg.id, m =>
			m.applyModeration(msg))

	handlers[message.stoleImageFrom] = (id: number) =>
		handle(id, m => {
			if (m.image) {
				m.image = null;

				m.view.removeImage();
				if (m instanceof FormModel) {
					m.allocatingImage = false;

					if (m.view && (m.view as any).upload) {
						(m.view as any).upload.reset();
					}
				}
				decrementImageCount();
			}
		})
	handlers[message.claudeAppend] = (message: ArrayBuffer) => {
		const view = new DataView(message);
		let id = view.getFloat64(0, true);
		let append = decoder.decode(message.slice(8));
		if(debug)
			console.log(`>binary claude append: ${id} ${append}`)
		handle(id,(m) => m.claudeAppend(append))
	}
	handlers[message.claudeDone] = (message: ArrayBuffer) => {
		const view = new DataView(message);
		let id = view.getFloat64(0, true);
		if(debug)
			console.log(`>binary claude done: ${id}`)
		handle(id,(m) => m.claudeDone())
	}

	interface StolenImage {
		id: number;
		image: ImageData;
	}

	handlers[message.stoleImageTo] = ({ id, image }: StolenImage) =>
		handle(id, m => {
			if (m.image) {
				m.image = null;
				m.view.removeImage();
			} else {
				incrementPostCount(false, true);
			}
			m.insertImage(image);
		})

	handlers[message.redirect] = (msg: string) => {
		const url = new URL(msg, location.origin)
		if (/https?:/.test(url.protocol)) {
			location.href = url.href
		}
	}

	handlers[message.notification] = (text: string) =>
		new OverlayNotification(text)

	handlers[message.setCookie] = ({ key, value }: CookieMessage) =>
		setCookie(key, value, 30)
}
