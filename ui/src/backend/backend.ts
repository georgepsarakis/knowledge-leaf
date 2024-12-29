import {parse} from "tldts";

export function getDomain() {
    const domain = window.location.hostname;
    if (domain === "localhost") {
        return 'http://' + domain + ":4000";
    }
    const tld = parse(window.location.hostname)
    return 'https://api.' + tld.domain;
}
