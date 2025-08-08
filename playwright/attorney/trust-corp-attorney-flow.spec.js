import {expect, test} from '@playwright/test';
import {randomAccessCode, screenshot, TestEmail} from '../e2e.js';
import {extractTextFromMainAndSave} from "../textExtractor.js";

test('trust corporation wants to be an attorney', async ({ page }) => {
    const accessCode = randomAccessCode()

    await page.goto(`/fixtures/attorney?redirect=&lpa-type=property-and-affairs&lpa-language=en&progress=signedByCertificateProvider&withAccessCode=${accessCode}&email=${TestEmail}&options=is-trust-corporation`);

    await page.goto(`/attorney-start`);
    await page.getByRole('button', { name: /Start|Start now/ }).click();
    await page.getByRole('button', { name: 'Continue' }).click();
    await page.getByRole('textbox', { name: 'Enter code' }).fill(accessCode);
    await page.getByRole('button', { name: 'Save and continue' }).click();

    await expect(page.getByLabel('Success').getByRole('paragraph')).toContainText('We have identified the trust corporation’s attorney access code');
    await screenshot(page)
    await extractTextFromMainAndSave(page)
    await page.getByRole('link', { name: 'Continue' }).click();
    await page.getByRole('link', { name: 'Confirm your details' }).click();

    await expect(page.locator('h1')).toContainText('What is your phone number?');
    await screenshot(page)
    await extractTextFromMainAndSave(page)
    await page.getByRole('button', { name: 'Save and continue' }).click();

    await expect(page.locator('h1')).toContainText('Your preferred language');
    await page.getByRole('radio', { name: 'English' }).check();
    await screenshot(page)
    await extractTextFromMainAndSave(page)
    await page.getByRole('button', { name: 'Save and continue' }).click();

    await expect(page.locator('h1')).toContainText('Confirm your trust corporation details');
    await screenshot(page)
    await extractTextFromMainAndSave(page)
    await page.getByRole('button', { name: 'Continue' }).click();
    await page.getByRole('link', { name: 'Read the LPA' }).click();

    await expect(page.locator('h1')).toContainText('Read the LPA');
    await screenshot(page)
    await extractTextFromMainAndSave(page)
    await page.getByRole('button', { name: 'Continue' }).click();
    await page.getByRole('link', { name: 'Sign the LPA' }).click();

    await expect(page.locator('h1')).toContainText('Legal rights and responsibilities');
    await screenshot(page)
    await extractTextFromMainAndSave(page)
    await page.getByRole('link', { name: 'Continue' }).click();

    await expect(page.locator('h1')).toContainText('What happens when you sign the LPA');
    await screenshot(page)
    await extractTextFromMainAndSave(page)
    await page.getByRole('link', { name: 'Continue to signing page' }).click();

    await expect(page.locator('h1')).toContainText('Sign the LPA on behalf of the trust corporation');
    await page.getByRole('textbox', { name: 'First names' }).fill('Authorised');
    await page.getByRole('textbox', { name: 'Last name' }).fill('Signatory');
    await page.getByRole('textbox', { name: 'Professional title' }).fill('One');
    await page.getByRole('checkbox', { name: 'I am authorised to sign on' }).check();
    await screenshot(page)
    await extractTextFromMainAndSave(page)
    await page.getByRole('button', { name: 'Submit signature' }).click();

    await expect(page.locator('h1')).toContainText('Would you like to add a second signatory?');
    await page.getByRole('radio', { name: 'Yes' }).check();
    await screenshot(page)
    await extractTextFromMainAndSave(page)
    await page.getByRole('button', { name: 'Continue' }).click();

    await expect(page.locator('h1')).toContainText('Sign the LPA on behalf of the trust corporation');
    await page.getByRole('textbox', { name: 'First names' }).fill('Authorised');
    await page.getByRole('textbox', { name: 'Last name' }).fill('Signatory');
    await page.getByRole('textbox', { name: 'Professional title' }).fill('Two');
    await page.getByRole('checkbox', { name: 'I am authorised to sign on' }).check();
    await screenshot(page)
    await extractTextFromMainAndSave(page)
    await page.getByRole('button', { name: 'Submit signature' }).click();

    await expect(page.locator('h1')).toContainText('Signing complete');
    await screenshot(page)
    await extractTextFromMainAndSave(page)
    await page.getByRole('link', { name: 'Return to ‘Manage LPAs’' }).click();
});
