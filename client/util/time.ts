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

export function relativeTimeAbbreviated(t: number): string {
    const currentTime = Math.floor(Date.now() / 1000);
    let timeDelta = Math.floor(currentTime - t),
        o = false;
    if (timeDelta < 1) {
        if (timeDelta > -50) return "now";
        o = true, timeDelta = -timeDelta
    } else if (timeDelta < 5) return "now";
    let i = "";
    const divide = [60, 60, 24, 30, 12],
        unit = ["s", "m", "h", "d", "mo"];
    for (let e = 0; e < divide.length; e++) {
        if (timeDelta < divide[e]) {
            i = `${timeDelta}${unit[e]}`;
            break
        }
        timeDelta = Math.floor(timeDelta / divide[e])
    }
    return "" === i && (i = `${timeDelta}y`), o && (i = "in " + i), i
}
