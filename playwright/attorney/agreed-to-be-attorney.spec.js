import {expect, test} from '@playwright/test';
import {randomShareCode, screenshot, TestEmail} from '../e2e.js';
import {extractTextFromMainAndSave} from "../textExtractor.js";

test('attorney agrees to be an attorney', async ({page}) => {
    const shareCode = randomShareCode()

    await page.goto(`/fixtures/attorney?redirect=&lpa-type=property-and-affairs&lpa-language=en&progress=signedByCertificateProvider&withShareCode=${shareCode}&email=${TestEmail}`);

    await page.goto(`/attorney-start`);
    await expect(page.locator('#main-content')).toContainText('Agree to be an attorney');
    await screenshot(page)
    await extractTextFromMainAndSave(page)
    await page.getByRole('button', {name: /Start|Start now/}).click();
    await page.getByRole('button', {name: 'Continue'}).click();

    await expect(page.locator('h1')).toContainText('Add an LPA');
    await page.getByRole('textbox', {name: 'Enter code'}).fill(shareCode);
    await screenshot(page)
    await extractTextFromMainAndSave(page)
    await page.getByRole('button', {name: 'Save and continue'}).click();

    await expect(page.getByLabel('Success').getByRole('paragraph')).toContainText('We have identified');
    await screenshot(page)
    await extractTextFromMainAndSave(page)
    await page.getByRole('link', {name: 'Continue'}).click();
    await page.getByRole('link', {name: 'Confirm your details'}).click();

    await expect(page.locator('h1')).toContainText('What is your phone number?');
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
    await page.getByRole('link', {name: 'Read the LPA'}).click();

    await expect(page.locator('h1')).toContainText('Read the LPA');
    await screenshot(page)
    await extractTextFromMainAndSave(page)
    await page.getByRole('button', {name: 'Continue'}).click();
    await page.getByRole('link', {name: 'Sign the LPA'}).click();

    await expect(page.locator('h1')).toContainText('Your legal rights and responsibilities');
    await screenshot(page)
    await extractTextFromMainAndSave(page)
    await page.getByRole('link', {name: 'Continue'}).click();

    await expect(page.locator('h1')).toContainText('What happens when you sign the LPA');
    await screenshot(page)
    await extractTextFromMainAndSave(page)
    await page.getByRole('link', {name: 'Continue to signing page'}).click();

    await expect(page.locator('h1')).toContainText('Sign as an attorney on this LPA');
    await page.getByRole('checkbox', {name: 'I, Jessie Jones, confirm'}).check();
    await screenshot(page)
    await extractTextFromMainAndSave(page)
    await page.getByRole('button', {name: 'Submit signature'}).click();

    await expect(page.locator('h1')).toContainText('You’ve formally agreed to be an attorney');
    await screenshot(page)
    await extractTextFromMainAndSave(page)
    await page.getByRole('link', {name: 'Go to your dashboard'}).click();
});
