import {expect, test} from '@playwright/test';
import {screenshot} from '../e2e.js'
import {extractTextFromMainAndSave} from "../textExtractor.js";

test('donor appoints replacement attorneys and chooses how they step in', async ({page}) => {
    await page.goto('/fixtures');
    await page.getByRole('radio', {name: 'Add a correspondent'}).check();
    await page.getByRole('button', {name: /Start|Start now/}).click();
    await page.getByRole('link', {name: 'Go to task list'}).click();
    await page.getByRole('link', {name: 'Choose your attorneys'}).click();
    await page.getByRole('radio', {name: 'No'}).check();
    await page.getByRole('button', {name: 'Continue'}).click();

    await expect(page.locator('h1')).toContainText('How your attorneys should make decisions');
    await page.getByRole('radio', {name: 'Jointly', exact: true}).check();
    await screenshot(page)
    await extractTextFromMainAndSave(page)
    await page.getByRole('button', {name: 'Save and continue'}).click();

    await expect(page.locator('h1')).toContainText('Because you have chosen for your attorneys to act jointly');
    await screenshot(page)
    await extractTextFromMainAndSave(page)
    await page.getByRole('link', {name: 'Return to task list'}).click();
    await page.getByRole('link', {name: 'Choose your replacement'}).click();
    await page.getByRole('radio', {name: 'No'}).check();
    await page.getByRole('button', {name: 'Continue'}).click();

    await expect(page.locator('h1')).toContainText('How should your replacement attorneys make decisions?');
    await page.getByRole('radio', {name: 'Jointly - your replacement'}).check();
    await screenshot(page)
    await extractTextFromMainAndSave(page)
    await page.getByRole('button', {name: 'Save and continue'}).click();
    await page.getByRole('link', {name: 'Check and send to your'}).click();

    await expect(page.locator('h1')).toContainText('Check your LPA');
    await screenshot(page)
    await extractTextFromMainAndSave(page)
});
