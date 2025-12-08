// noinspection JSUnusedGlobalSymbols

import { parse } from "csv-parse/sync";
import { stringify } from "csv-stringify/sync";
import fs from "fs";
import { fetchGbTransactions } from "./fetchGbData";
import { writeFileSync } from "node:fs";

interface Transaction {
    date: string; // ISO 8601 date string in PST.
    donation: string;
    donor: string;
    matchName: string;
}

type TransactionsByDate = { [date: string]: Transaction[] };

type TransactionsByAmount = { [donation: string]: Transaction[] };

main();

function main() {
    loadTransactions().then(({ ats, gbs }) => {
        createCsvFiles("initial", ats, gbs);
        reconcileTransactions(ats, gbs).then(({ atu, gbu }) => {
            createCsvFiles("reconciled", atu, gbu);
        });
    });
}

export async function loadTransactions() {
    let ats = loadAirtableTransactions();
    let gbs = await loadGbTransactions(ats[0].date);
    return { ats, gbs };
}

export function createCsvFiles(prefix: string, ats: Transaction[], gbs: Transaction[]) {
    const composite = createComposite(ats, gbs);
    writeFileSync(`../../local/composite-donations.csv`, stringify(composite));
    const atuCsv = stringify(ats, {
        header: true,
        columns: ["date", "donation", "donor", "matchName"],
    });
    writeFileSync(`../../local/${prefix}-airtable-donations.csv`, atuCsv);
    const gbuCsv = stringify(gbs, {
        header: true,
        columns: ["date", "donation", "donor", "matchName"],
    });
    writeFileSync(`../../local/${prefix}-givebutter-donations.csv`, gbuCsv);
}

export async function reconcileTransactions(ats: Transaction[], gbs: Transaction[]) {
    const gbCandidates: TransactionsByDate = organizeByDate(gbs);
    let checkForMatch = (at: Transaction, date: string) => {
        const gbd = gbCandidates[date] ?? [];
        for (let g = 0; g < gbd.length; g++) {
            if (amountAndDonorMatch(at, gbd[g])) {
                gbd.splice(g, 1);
                return true;
            }
        }
        return false;
    };
    for (let a = 0; a < ats.length; ) {
        const at = ats[a];
        if (checkForMatch(at, at.date)) {
            ats.splice(a, 1);
            continue;
        }
        if (checkForMatch(at, dayBefore(at.date))) {
            ats.splice(a, 1);
            continue;
        }
        a++;
    }
    console.log(`Airtable: ${ats.length} unreconciled transactions`);
    const gbu = deOrganize(gbCandidates);
    console.log(`GiveButter: ${gbu.length} unreconciled transactions`);
    return { atu: ats, gbu };
}

function loadAirtableTransactions() {
    let path = "../../local/airtable-donations.csv";
    let content = fs.readFileSync(path, "utf8");
    let ts: Transaction[] = parse(content, { columns: true });
    ts = ts.map(
        (t) =>
            ({
                donation: t.donation,
                date: t.date,
                donor: t.donor,
                matchName: findLastName(t.donor),
            }) as Transaction
    );
    console.log(`Airtable: ${ts.length} transactions`);
    return ts.sort(orderTransactions);
}

async function loadGbTransactions(latestDate: string | undefined = undefined) {
    let transactions = await fetchGbTransactions();
    const converter = new Intl.DateTimeFormat("sv", { timeZone: "America/Los_Angeles" });
    const summaries: Transaction[] = [];
    for (const t of transactions) {
        if (!t.payout) {
            continue;
        }
        const date = converter.format(new Date(t.created_at)).slice(0, 10);
        if (latestDate && date <= latestDate) {
            summaries.push({
                donation: t.amount.toString(),
                date,
                donor: t.first_name + " " + t.last_name,
                matchName: findLastName(t.first_name + " " + t.last_name),
            } as Transaction);
        }
    }
    if (latestDate) {
        console.log(`GiveButter: ${summaries.length} transactions before ${latestDate}`);
    } else {
        console.log(`GiveButter: ${summaries.length} transactions`);
    }
    return summaries.sort(orderTransactions);
}

export function organizeByDate(transactions: Transaction[]) {
    return transactions.reduce((acc, summary) => {
        const date = summary.date;
        acc[date] = acc[date] || [];
        acc[date].push(summary);
        return acc;
    }, {} as TransactionsByDate);
}

export function organizeByAmount(transactions: Transaction[]) {
    return transactions.reduce((summary, transaction) => {
        const donation = transaction.donation;
        summary[donation] = summary[donation] || [];
        summary[donation].push(transaction);
        return summary;
    }, {} as TransactionsByAmount);
}

export function deOrganize(groupedTransactions: { [key: string]: Transaction[] }) {
    const all = Object.values(groupedTransactions).flat();
    return all.sort(orderTransactions);
}

function orderTransactions(s1: Transaction, s2: Transaction) {
    if (s1.date === s2.date) {
        if (s1.donation === s2.donation) {
            return s2.matchName.localeCompare(s1.matchName);
        } else {
            return Number(s2.donation) - Number(s1.donation);
        }
    } else {
        return s2.date.localeCompare(s1.date);
    }
}

function amountAndDonorMatch(t1: Transaction, t2: Transaction) {
    return t1.donation === t2.donation && t1.matchName === t2.matchName;
}

function dayBefore(date: string) {
    const d = new Date(date);
    d.setDate(d.getDate() - 1);
    return d.toISOString().slice(0, 10);
}

function findLastName(name: string) {
    const parts = name.split(/[- ]+/).reverse();
    for (let part of parts) {
        part = part.replace(/\W/g, "");
        // remove suffixes like "Jr" and "MD"
        if (part.length <= 2) {
            continue;
        }
        return part.toLowerCase();
    }
    return parts[0].toLowerCase().replace(/\W/g, "");
}

function createComposite(ats: Transaction[], gbs: Transaction[]) {
    // create composite records with Date, Airtable Amount, Airtable Donor, GiveButter Amount, GiveButter Donor
    ats.sort(orderTransactions);
    gbs.sort(orderTransactions);
    const result = [
        ["Date", "Airtable Amount", "Airtable Donor", "GiveButter Amount", "GiveButter Donor"],
    ];
    for (let a = 0, g = 0; a < ats.length && g < gbs.length; ) {
        if (ats[a].date === gbs[g].date) {
            result.push([
                ats[a].date,
                ats[a].donation,
                ats[a].donor,
                gbs[g].donation,
                gbs[g].donor,
            ]);
            a++;
            g++;
        } else if (ats[a].date > gbs[g].date) {
            result.push([ats[a].date, ats[a].donation, ats[a].donor, "", ""]);
            a++;
        } else {
            result.push([gbs[g].date, "", "", gbs[g].donation, gbs[g].donor]);
            g++;
        }
    }
    return result;
}
