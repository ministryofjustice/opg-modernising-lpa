import {expect, test} from '@playwright/test';
import {screenshot} from '../e2e.js'
import {extractTextFromMainAndSave} from "../textExtractor.js";

test('certificate provider opts out of being a certificate provider (at declaration)', async ({page}) => {
    await page.goto('/fixtures');
    await page.getByRole('link', {name: 'Certificate provider'}).click();
    await page.getByRole('radio', {name: 'Confirm your identity'}).check();
    await page.getByRole('button', {name: /Start|Start now/}).click();
    await page.getByRole('tab', {name: 'I’m a certificate provider'}).click();
    await page.getByRole('link', {name: 'Go to task list'}).click();
    await page.getByRole('link', {name: 'Provide your certificate'}).click();
    await page.getByRole('button', {name: 'Continue'}).click();
    await page.getByRole('link', {name: 'Continue'}).click();
    await page.getByRole('button', {name: 'I cannot provide the'}).click();

    await expect(page.locator('h1')).toContainText('Confirm you do not want to be the certificate provider');
    await screenshot(page)
    await extractTextFromMainAndSave(page)
    await page.getByRole('button', {name: 'Confirm'}).click();

    await expect(page.locator('h1')).toContainText('You have confirmed that you do not want to be Sam Smith’s certificate provider.');
    await screenshot(page)
    await extractTextFromMainAndSave(page)
});
