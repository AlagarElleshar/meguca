import {VideoItem, VideoItemList} from "../typings/messages";


export class VideoList {
    public items: VideoItem[] = [];
    public pos: number = 0;
    public isOpen: boolean = true;

    public get length(): number {
        return this.items.length;
    }

    public get currentItem(): VideoItem | null {
        return this.items[this.pos] ?? null;
    }

    public getItem(i: number): VideoItem {
        return this.items[i];
    }

    public setItem(i: number, item: VideoItem): void {
        this.items[i] = item;
    }

    public getItems(): VideoItem[] {
        return this.items;
    }

    public setItems(items: VideoItem[]): void {
        this.items = items
        this.pos = 0;
    }

    public setPos(i: number): void {
        if (i < 0 || i > this.length - 1) i = 0;
        this.pos = i;
    }

    public exists(f: (item: VideoItem) => boolean): boolean {
        return this.items.some(f);
    }

    public findIndex(f: (item: VideoItem) => boolean): number {
        return this.items.findIndex(f);
    }

    public addItem(item: VideoItem, atEnd: boolean): void {
        if (atEnd) this.items.push(item);
        else this.items.splice(this.pos + 1, 0, item);
    }

    public setNextItem(nextPos: number): void {
        const next = this.items[nextPos];
        this.items.splice(nextPos, 1);
        if (nextPos < this.pos) this.pos--;
        this.items.splice(this.pos + 1, 0, next);
    }

    public removeItem(index: number): void {
        if (index < this.pos) this.pos--;
        this.items.splice(index, 1);
        if (this.pos >= this.items.length) this.pos = 0;
    }

    public skipItem(): void {
        const item = this.items[this.pos];
        if (!item.isTemp) this.pos++;
        else this.items.splice(this.pos, 1);
        if (this.pos >= this.items.length) this.pos = 0;
    }

    public clear(): void {
        this.items = [];
        this.pos = 0;
    }
}