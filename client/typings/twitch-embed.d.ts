declare namespace Twitch {
  interface PlayerOptions {
    width?: number | string;
    height?: number | string;
    channel?: string;
    video?: string;
    collection?: string;
    autoplay?: boolean;
    muted?: boolean;
    time?: string;
    parent?: string[];
  }

  interface PlaybackStats {
    backendVersion: string;
    bufferSize: number;
    codecs: string;
    displayResolution: string;
    fps: number;
    hlsLatencyBroadcaster: number;
    playbackRate: number;
    skippedFrames: number;
    videoResolution: string;
  }

  class Player {
    constructor(id: string, options: PlayerOptions);
    disableCaptions(): void;
    enableCaptions(): void;
    pause(): void;
    play(): void;
    seek(timestamp: number): void;
    setChannel(channel: string): void;
    setCollection(collectionID: string, videoID?: string): void;
    setQuality(quality: string): void;
    setVideo(videoID: string, timestamp?: number): void;
    getMuted(): boolean;
    setMuted(muted: boolean): void;
    getVolume(): number;
    setVolume(volumeLevel: number): void;
    getPlaybackStats(): PlaybackStats;
    getChannel(): string;
    getCurrentTime(): number;
    getDuration(): number;
    getEnded(): boolean;
    getQualities(): string[];
    getQuality(): string;
    getVideo(): string;
    isPaused(): boolean;
    addEventListener(event: string, callback: (data: any) => void): void;
    removeEventListener(event: string, callback: (data: any) => void): void;
  }

  namespace Player {
    const CAPTIONS: string;
    const ENDED: string;
    const PAUSE: string;
    const PLAY: string;
    const PLAYBACK_BLOCKED: string;
    const PLAYING: string;
    const OFFLINE: string;
    const ONLINE: string;
    const READY: string;
    const SEEK: string;
  }
}