import {RawPlayer} from "./rawplayer";
import {VideoItem} from "../../typings/nekotv";
import {isCuck} from "../../common";

export class TikTokPlayer extends RawPlayer {
    override initMediaPlayer() {
        super.initMediaPlayer();
        this.videoElement.id = 'tiktok-player';
    }

    override loadVideo(item: VideoItem): void {
        this.loaded = false;
        if(this.videoElement == null) {
            this.initMediaPlayer()
        }
        if(isCuck){
            this.videoElement.src = `https://tikwm.com/video/media/play/${item.id}.mp4`
        }
        else{
            this.videoElement.src = `https://tikwm.com/video/media/hdplay/${item.id}.mp4`
        }
    }
}