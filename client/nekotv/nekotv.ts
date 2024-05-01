import {message, sendBinary} from "../connection";

export function initNekoTV(){
    let nekoTV = document.getElementById("banner-nekotv")
    if (!nekoTV){
        return;
    }
    nekoTV.addEventListener("click", () => {
        sendBinary(new Uint8Array([message.nekoTV]));
    })
}