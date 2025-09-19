import {expect, test} from '@playwright/test';
import {screenshot} from '../e2e.js'
import {extractTextFromMainAndSave} from "../textExtractor.js";

test('voucher cannot confirm their identity', async ({page}) => {
    await page.goto('/fixtures');
    await page.getByRole('link', {name: 'Voucher'}).click();
    await page.getByRole('radio', {name: 'Verify donor details'}).check();
    await page.getByRole('button', {name: /Start|Start now/}).click();
    await page.getByRole('link', {name: 'Go to task list'}).click();
    await page.getByRole('link', {name: 'Confirm your identity'}).click();
    await page.getByRole('button', {name: 'Continue'}).click();
    await page.getByRole('radio', {name: 'Failed identity check (T)'}).check();
    await page.getByRole('button', {name: 'Continue'}).click();

    await expect(page.getByLabel('Important').getByRole('paragraph')).toContainText('You were not able to confirm your identity using GOV.UK One Login.');
    await screenshot(page)
    await extractTextFromMainAndSave(page)
});
