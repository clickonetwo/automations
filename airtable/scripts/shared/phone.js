/*
 * Copyright 2024-2025 Daniel C. Brotsky. All rights reserved.
 * All the copyrighted work in this repository is licensed under the
 * open source MIT License, reproduced in the LICENSE file.
 */

// takes a phone in E.164 format and formats it for display
function formatPhone(phone) {
    const twoDigitCountryCodes = [
        "20", "27", "30", "31", "32", "33", "34", "36", "39",
        "40", "41", "43", "44", "45", "46", "47", "48", "49",
        "51", "52", "53", "54", "55", "56", "57", "58",
        "60", "61", "62", "63", "64", "65", "66",
        "70", "71", "72", "73", "74", "75", "76", "77", "78", "79",
        "81", "82", "84", "86", "87", "88",
        "90", "91", "92", "93", "94", "95", "98",
    ]
    if (phone.startsWith("+1")) {
        // Zone 1
        if (phone.length < 12) {
            return "invalid"
        }
        const part1 = `(${phone.substring(2,5)}) ${phone.substring(5,8)}-${phone.substring(8,12)}`
        if (phone.length === 12) {
            return part1
        } else {
            return part1 + " x" + phone.substring(12)
        }
    }
    // international
    let prefix = phone.substring(0, 3);
    let suffix = phone.substring(3);
    if (!twoDigitCountryCodes.includes(phone.substring(1, 3))) {
        prefix = phone.substring(0, 4);
        suffix = phone.substring(4);
    }
    let suffixSuffix = "";
    if (suffix.length % 3 === 1) {
        // put last four numbers together
        suffixSuffix = "-" + suffix.substring(suffix.length - 4);
        suffix = suffix.substring(0, suffix.length - 4);
    }
    for (let i = suffix.length; i > 3; i = i - 3) {
        suffix = suffix.substring(0, i-3) + "-" + suffix.substring(i-3)
    }
    return prefix + "-" + suffix + suffixSuffix;
}

// takes a flexibly-formatted number that starts with '+' and strips non-digits
function canonicalizePhone(phone) {
    if (!phone.startsWith("+")) {
        throw new Error(`Form phone number '${phone}' is invalid.`)
    }
    let digits = phone.replace(/\D/g,'');
    return "+" + digits
}
