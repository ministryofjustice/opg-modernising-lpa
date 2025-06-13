import {expect, test} from '@playwright/test';
import {randomShareCode, screenshot, TestEmail} from '../../../test/playwright/e2e.js'
import {extractTextFromMainAndSave} from "../../../test/playwright/textExtractor.js";

test('certificate provider starts journey', async ({page}) => {
    const shareCode = randomShareCode()

    await page.goto(`/fixtures/certificate-provider?redirect=&lpa-type=property-and-affairs&lpa-language=en&progress=paid&withShareCode=${shareCode}&email=${TestEmail}`);

    await page.goto(`/certificate-provider-start`);

    await expect(page.locator('#main-content')).toContainText('Act as a certificate provider');
    await screenshot(page)
    await extractTextFromMainAndSave(page)
    await page.getByRole('button', {name: /Start|Start now/}).click();
    await page.getByRole('button', {name: 'Continue'}).click();
    await page.getByRole('textbox', {name: 'Enter code'}).fill(shareCode);

    await expect(page.locator('h1')).toContainText('Add an LPA');
    await screenshot(page)
    await extractTextFromMainAndSave(page)
    await page.getByRole('button', {name: 'Save and continue'}).click();

    await expect(page.getByLabel('Success').getByRole('paragraph')).toContainText('We have identified your certificate provider access code');
    await screenshot(page)
    await extractTextFromMainAndSave(page)
    await page.getByRole('link', {name: 'Continue'}).click();

    await expect(page.locator('h1')).toContainText('Enter your date of birth');
    await page.getByRole('textbox', {name: 'Day'}).fill('23');
    await page.getByRole('textbox', {name: 'Month'}).fill('01');
    await page.getByRole('textbox', {name: 'Year'}).fill('1992');
    await screenshot(page)
    await extractTextFromMainAndSave(page)
    await page.getByRole('button', {name: 'Save and continue'}).click();

    await expect(page.locator('h1')).toContainText('Your preferred language');
    await page.getByRole('radio', {name: 'English'}).check();
    await screenshot(page)
    await extractTextFromMainAndSave(page)
    await page.getByRole('button', {name: 'Save and continue'}).click();

    await expect(page.locator('h1')).toContainText('Confirm your details');
    await screenshot(page)
    await extractTextFromMainAndSave(page)
    await page.getByRole('button', {name: 'Continue'}).click();

    await expect(page.locator('h1')).toContainText('Your role as certificate provider');
    await screenshot(page)
    await extractTextFromMainAndSave(page)
    await page.getByRole('link', {name: 'Go to your task list'}).click();

    await expect(page.locator('h1')).toContainText('Your task list');
});
