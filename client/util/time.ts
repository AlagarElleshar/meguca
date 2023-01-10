// Time related aids
import lang from '../lang'

export function secondsToTime(s: number): string {
    const divide = [60, 60, 24, 30, 12]
    const unit = ['second', 'minute', 'hour', 'day', 'month']
    let time = s

    const format = (key: string) => {
        let tmp = time.toFixed(1)
        let plural = lang.plurals[key][1]

        if (tmp.includes(".0")) {
            tmp = tmp.substr(0, tmp.length - 2)

            if (tmp == '1') {
                plural = lang.plurals[key][0]
            }
        }

        return `${tmp} ${plural}`
    }

    for (let i = 0; i < divide.length; i++) {
        if (time < divide[i]) {
            return format(unit[i])
        }

        time /= divide[i]
    }

    return format("year")
}


export function timeDelta(e: number): number {
    const t = Math.floor(Date.now() / 1000);
    return Math.floor((t - e) / 60)
}

export function relativeTimeAbbreviated(e: number): string {
    const t = Math.floor(Date.now() / 1000);
    let s = Math.floor(t - e),
        o = false;
    if (s < 1) {
        if (s > -50) return "now";
        o = true, s = -s
    } else if (s < 5) return "now";
    let i = "";
    const n = [60, 60, 24, 30, 12],
        r = ["s", "m", "h", "d", "mo"];
    for (let e = 0; e < n.length; e++) {
        if (s < n[e]) {
            i = `${s}${r[e]}`;
            break
        }
        s = Math.floor(s / n[e])
    }
    return "" === i && (i = `${s}y`), o && (i = "in " + i), i
}

export function relativeSyncedTime(e: number, t: number): string {
    let s = Math.floor(t - e);
    if (s < 1) return "0s";
    let o = "";
    const i = [60, 60, 24, 30, 12],
        n = ["s", "m", "h", "d", "mo"];
    for (let e = 0; e < i.length; e++) {
        if (s < i[e]) {
            o = `${s}${n[e]}`;
            break
        }
        s = Math.floor(s / i[e])
    }
    return "" === o && (o = `${s}y`), o
}

// const unhideTimeElement = document.getElementById("unhide-time");
//
// export function unhideTimestamps(): void {
//     unhideTimeElement.innerHTML = ".hidden-time { visibility: visible; font-size: inherit; }"
// }
//
// export function hideTimestamps(): void {
//     unhideTimeElement.innerHTML = ""
// }
