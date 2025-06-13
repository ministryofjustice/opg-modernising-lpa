import {expect, test} from '@playwright/test';
import {screenshot} from '../e2e.js'
import {extractTextFromMainAndSave} from "../textExtractor.js";

test('voucher starts their journey', async ({page}) => {
    await page.goto('/fixtures');
    await page.getByRole('link', {name: 'Voucher'}).click();
    await page.getByRole('button', {name: /Start|Start now/}).click();

    await expect(page.locator('h1')).toContainText('What is vouching?');
    await screenshot(page)
    await extractTextFromMainAndSave(page)
    await page.getByRole('button', {name: /Start|Start now/}).click();
    await page.getByRole('button', {name: 'Continue'}).click();

    await expect(page.locator('h1')).toContainText('Vouch for someone');
    await screenshot(page)
    await extractTextFromMainAndSave(page)
});
