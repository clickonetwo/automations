// noinspection JSUnusedGlobalSymbols

import fs from "fs";
import { parse } from "csv-parse/sync";
import { GbTransactionPayload } from "./payloads";

export function convertCsvToJson() {
    const path = "../../local/development-transactions.";
    const input = fs.readFileSync(path + "csv", "utf8");
    let rows = parse(input, { columns: false });
    console.log(`Read ${rows.length} rows`);
    const payloads = [];
    for (const row of rows) {
        const payload = JSON.parse(row[0]);
        payloads.push(payload);
    }
    const output = JSON.stringify(payloads);
    fs.writeFileSync(path + "json", output);
    console.log(`Wrote ${payloads.length} payloads`);
}

export function loadPayloads() {
    const path = "../../local/development-transactions.";
    const content = fs.readFileSync(path + "json", "utf8");
    return JSON.parse(content) as GbTransactionPayload[];
}

export async function pushJsonToMake(payload: GbTransactionPayload) {
    // noinspection SpellCheckingInspection
    const signature =
        "IHCO60sgh7TJwpky3eojtQhy77VsA2zvznSDbhCH2vbwabU7bULi8hBJtM5aVqQVS6Xehs5T0QzgcPt5fqFDcyWCqCS9NkfJ1NIr";
    const endpoint = "https://hook.us1.make.com/i5qsl644dvk8m6jinfgno8ssx03b7lum";
    const result = await fetch(endpoint, {
        method: "POST",
        headers: {
            "Content-Type": "application/json",
            Signature: signature,
        },
        body: JSON.stringify(payload),
    });
    if (!result.ok) {
        throw new Error(`HTTP response ${result.status}: ${result.statusText}`);
    }
}

export async function pushPayloads(start = 0, end = 1000) {
    const payloads = loadPayloads();
    end = Math.min(end, payloads.length);
    for (let i = start; i < end; i++) {
        if (i > start) {
            console.log(`Waiting 15 seconds before next push...`);
            await new Promise((resolve) => setTimeout(resolve, 15000));
        }
        console.log(`Pushing payload ${i}...`);
        try {
            await pushJsonToMake(payloads[i]);
        } catch (e) {
            console.error(`Error pushing payload ${i}: ${e}`);
            break;
        }
    }
}
