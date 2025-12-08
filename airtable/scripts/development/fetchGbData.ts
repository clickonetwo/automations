// noinspection JSUnusedGlobalSymbols

import {
    GbTransactionData,
    GbListTransactionsPayload,
    GbPlanData,
    GbListPlansPayload,
    GbCampaignData,
    GbListCampaignsPayload,
} from "./payloads";
import fs from "fs";

async function fetchAllGbTransactions() {
    // noinspection SpellCheckingInspection
    const gbApiKey = "8513|QykGq6xF69yvSDsWsJG4fGq6OsrvLRwrG4TvW5vs";
    let gbApiEndpoint = `https://api.givebutter.com/v1/transactions`;
    let page = 1;
    let more = true;
    const transactions: GbTransactionData[] = [];
    while (more) {
        const response = await fetch(gbApiEndpoint + `?page=${page}`, {
            method: "GET",
            headers: {
                Accept: "application/json",
                Authorization: `Bearer ${gbApiKey}`,
            },
        });
        if (!response.ok) {
            throw new Error(
                `Fetch error on page ${page}: ${response.status}, ${response.statusText}`
            );
        }
        page++;
        const body: GbListTransactionsPayload = await response.json();
        const last_page = body.meta.last_page ?? 0;
        more = last_page > page;
        for (const payload of body.data) {
            transactions.push(payload);
        }
        console.log(
            `After page ${page} of ${last_page}, found ${transactions.length} transactions`
        );
    }
    return transactions;
}

export async function fetchGbTransactions() {
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

async function fetchAllGbPlans() {
    // noinspection SpellCheckingInspection
    const gbApiKey = "8513|QykGq6xF69yvSDsWsJG4fGq6OsrvLRwrG4TvW5vs";
    let gbApiEndpoint = `https://api.givebutter.com/v1/plans`;
    let page = 1;
    let more = true;
    const plans: GbPlanData[] = [];
    while (more) {
        const response = await fetch(gbApiEndpoint + `?page=${page}`, {
            method: "GET",
            headers: {
                Accept: "application/json",
                Authorization: `Bearer ${gbApiKey}`,
            },
        });
        if (!response.ok) {
            throw new Error(
                `Fetch error on page ${page}: ${response.status}, ${response.statusText}`
            );
        }
        page++;
        const body: GbListPlansPayload = await response.json();
        const last_page = body.meta.last_page ?? 0;
        more = last_page > page;
        for (const payload of body.data) {
            plans.push(payload);
        }
        console.log(`After page ${page} of ${last_page}, found ${plans.length} plans`);
    }
    return plans;
}

export async function fetchGbPlans() {
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

async function fetchAllGbCampaigns() {
    // noinspection SpellCheckingInspection
    const gbApiKey = "8513|QykGq6xF69yvSDsWsJG4fGq6OsrvLRwrG4TvW5vs";
    let gbApiEndpoint = `https://api.givebutter.com/v1/campaigns`;
    let page = 1;
    let more = true;
    const campaigns: GbCampaignData[] = [];
    while (more) {
        const response = await fetch(gbApiEndpoint + `?page=${page}`, {
            method: "GET",
            headers: {
                Accept: "application/json",
                Authorization: `Bearer ${gbApiKey}`,
            },
        });
        if (!response.ok) {
            throw new Error(
                `Fetch error on page ${page}: ${response.status}, ${response.statusText}`
            );
        }
        page++;
        const body: GbListCampaignsPayload = await response.json();
        const last_page = body.meta.last_page ?? 0;
        more = last_page > page;
        for (const payload of body.data) {
            campaigns.push(payload);
        }
        console.log(`After page ${page} of ${last_page}, found ${campaigns.length} campaigns`);
    }
    return campaigns;
}

export async function fetchGbCampaigns() {
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
