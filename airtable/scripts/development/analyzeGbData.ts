// noinspection JSUnusedGlobalSymbols

import { fetchGbTransactions } from "./fetchGbData";

export async function findTransactionsWithMismatchedAmounts() {
    const transactions = await fetchGbTransactions();
    const mismatches = transactions.filter((d) => d.amount !== d.donated);
    const mismatchesCoveredFee = mismatches.filter((d) => d.fee_covered == d.fee);
    const ms = mismatches.map((d) => ({
        id: d.id,
        amount: d.amount,
        fee: d.fee,
        fee_covered: d.fee_covered,
        donated: d.donated,
        payout: d.payout,
    }));
    const msCoveredFee = mismatchesCoveredFee.map((d) => ({
        id: d.id,
        amount: d.amount,
        donated: d.donated,
        payout: d.payout,
    }));
    return { ms, msCoveredFee };
}

export async function findGbCustomFieldTitles() {
    const donations = await fetchGbTransactions();
    const customFields = donations
        .filter((d) => d.custom_fields.length)
        .map((d) => d.custom_fields);
    const titles: string[] = [];
    for (const fields of customFields) {
        for (const field of fields) {
            if (!titles.includes(field.title)) {
                titles.push(field.title);
            }
        }
    }
    return titles;
}
