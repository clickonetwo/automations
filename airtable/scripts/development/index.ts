// noinspection JSUnusedLocalSymbols

import fs from "fs";
import { fetchAllGbCampaigns, fetchAllGbPlans, fetchAllGbTransactions } from "./fetchGbDataLocally";
import { GbCampaignData, GbPlanData, GbTransactionData } from "./payloads";

async function fetchGbDonations() {
    let path = "../../local/donations.json";
    if (!fs.existsSync(path)) {
        console.log("No local donations.json file found, fetching from GB API...");
        const donations = await fetchAllGbTransactions();
        let content = JSON.stringify(donations, null, 2);
        fs.writeFileSync(path, content);
        console.log(`Wrote ${donations.length} donations to local donations.json file.`);
        return donations;
    }
    let donations: GbTransactionData[] = JSON.parse(fs.readFileSync(path, "utf8"));
    return donations;
}

async function fetchGbPlans() {
    let path = "../../local/plans.json";
    if (!fs.existsSync(path)) {
        console.log("No local plans.json file found, fetching from GB API...");
        const plans = await fetchAllGbPlans();
        let content = JSON.stringify(plans, null, 2);
        fs.writeFileSync(path, content);
        console.log(`Wrote ${plans.length} plans to local plans.json file.`);
        return plans;
    }
    let plans: GbPlanData[] = JSON.parse(fs.readFileSync(path, "utf8"));
    return plans;
}

async function fetchGbCampaigns() {
    let path = "../../local/campaigns.json";
    if (!fs.existsSync(path)) {
        console.log("No local campaigns.json file found, fetching from GB API...");
        const campaigns = await fetchAllGbCampaigns();
        let content = JSON.stringify(campaigns, null, 2);
        fs.writeFileSync(path, content);
        console.log(`Wrote ${campaigns.length} campaigns to local campaigns.json file.`);
        return campaigns;
    }
    let campaigns: GbCampaignData[] = JSON.parse(fs.readFileSync(path, "utf8"));
    return campaigns;
}

async function findDonationsWithMismatchedAmounts() {
    const donations = await fetchGbDonations();
    const mismatches = donations.filter((d) => d.amount !== d.donated);
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

async function findGbCustomFieldTitles() {
    const donations = await fetchGbDonations();
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

function main() {
    fetchGbCampaigns().then((campaigns) => {
        for (const c of campaigns) {
            console.log(`${c.id}, ${c.code}, ${c.title}`);
        }
    });
}

main();
