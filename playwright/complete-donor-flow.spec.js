import {expect, test} from '@playwright/test';
import {screenshot} from './e2e';
import {extractTextFromMainAndSave} from "./textExtractor";

test('donor happy path', async ({ page }) => {
    await page.goto('http://localhost:5050/start');
    await page.getByRole('button', { name: 'Start' }).click();

    await page.getByRole('button', { name: 'Continue' }).click();

    await expect(page).toHaveURL(/\/dashboard/);
    await page.getByRole('button', { name: 'Accept analytics cookies' }).click();
    await screenshot(page)
    await extractTextFromMainAndSave(page)
    await page.getByRole('button', { name: 'Start now' }).click();

    await expect(page).toHaveURL(/\/your-name/);
    await page.getByLabel('First names').fill('Samuel');
    await page.getByLabel('Last name').fill('Smith');
    await screenshot(page)
    await extractTextFromMainAndSave(page)
    await page.getByRole('button', { name: 'Save and continue' }).click();

    await expect(page).toHaveURL(/\/your-date-of-birth/);
    await page.getByLabel('Day').fill('01');
    await page.getByLabel('Month').fill('01');
    await page.getByLabel('Year').fill('2000');
    await screenshot(page)
    await extractTextFromMainAndSave(page)
    await page.getByRole('button', { name: 'Save and continue' }).click();

    await expect(page).toHaveURL(/\/do-you-live-in-the-uk/);
    await page.getByRole('group', { name: 'Do you live in the UK, the Channel Islands or the Isle of Man?' }).getByLabel('Yes').check();
    await screenshot(page)
    await extractTextFromMainAndSave(page)
    await page.getByRole('button', { name: 'Continue' }).click();

    await expect(page).toHaveURL(/\/your-address/);
    await page.getByLabel('Postcode').fill('B14 7ED');
    await screenshot(page)
    await extractTextFromMainAndSave(page)
    await page.getByLabel('Postcode').press('Enter');

    await expect(page).toHaveURL(/\/your-address/);
    await page.getByLabel('Select an address').selectOption('{"line1":"1 RICHMOND PLACE","line2":"","line3":"","town":"BIRMINGHAM","postcode":"B14 7ED","country":"GB"}');
    await screenshot(page)
    await extractTextFromMainAndSave(page)
    await page.getByRole('button', { name: 'Continue' }).click();

    await expect(page).toHaveURL(/\/your-address/);
    await screenshot(page)
    await extractTextFromMainAndSave(page)
    await page.getByRole('button', { name: 'Save and continue' }).click();

    await expect(page).toHaveURL(/\/receiving-updates-about-your-lpa/);
    await page.getByText('If you do not want to be contacted about your LPA').click();
    await screenshot(page)
    await extractTextFromMainAndSave(page)
    await page.getByRole('link', { name: 'Continue' }).click();

    await expect(page).toHaveURL(/\/can-you-sign-your-lpa/);
    await page.getByLabel('Yes').check();
    await screenshot(page)
    await extractTextFromMainAndSave(page)
    await page.getByRole('button', { name: 'Save and continue' }).click();

    await expect(page).toHaveURL(/\/your-preferred-language/);
    await page.getByText('How to change the language to Welsh').click();
    await page.getByRole('group', { name: 'Which language would you prefer us to use when we contact you?' }).getByLabel('English').check();
    await page.getByRole('group', { name: 'In which language would you' }).getByLabel('English').check();
    await screenshot(page)
    await extractTextFromMainAndSave(page)
    await page.getByRole('button', { name: 'Save and continue' }).click();

    await expect(page).toHaveURL(/\/your-legal-rights-and-responsibilities-if-you-make-an-lpa/);
    await screenshot(page)
    await extractTextFromMainAndSave(page)
    await page.getByRole('link', { name: 'Continue' }).click();

    await expect(page).toHaveURL(/\/lpa-type/);
    await page.getByLabel('Property and affairs').check();
    await screenshot(page)
    await extractTextFromMainAndSave(page)
    await page.getByRole('button', { name: 'Save and continue' }).click();

    await expect(page).toHaveURL(/\/task-list/);

    // Give UID service time to return UID
    await expect.poll(async () => {
        await page.reload();
        const referenceNumber = page.getByText('M-');
        return await referenceNumber.isVisible() && (await referenceNumber.innerText()) !== "";
    }, {
        timeout: 20000,
        intervals: [500]
    }).toBe(true);

    await page.getByText('Help with making this LPA').click();
    await screenshot(page)
    await extractTextFromMainAndSave(page)
    await page.getByRole('link', { name: 'Choose your attorneys' }).click();

    await expect(page).toHaveURL(/\/choose-attorneys-guidance/);
    await page.getByText('Trust corporations', { exact: true }).click();
    await screenshot(page)
    await extractTextFromMainAndSave(page)
    await page.getByRole('button', { name: 'Continue' }).click();

    await expect(page).toHaveURL(/\/choose-attorneys/);
    await page.getByText('Guidance on choosing an attorney').click();
    await page.getByLabel('First names').fill('Test');
    await page.getByLabel('Last name').fill('Testington');
    await page.getByLabel('Day').fill('01');
    await page.getByLabel('Month').fill('0');
    await page.getByLabel('Month').fill('01');
    await page.getByLabel('Year').fill('2000');
    await screenshot(page)
    await extractTextFromMainAndSave(page)
    await page.getByRole('button', { name: 'Save and continue' }).click();

    await expect(page).toHaveURL(/\/choose-attorneys-address/);
    await page.getByLabel('Use an address you’ve already').check();
    await screenshot(page)
    await extractTextFromMainAndSave(page)
    await page.getByRole('button', { name: 'Continue' }).click();

    await expect(page).toHaveURL(/\/choose-attorneys-address/);
    await page.getByText('1 RICHMOND PLACE').click();
    await screenshot(page)
    await extractTextFromMainAndSave(page)
    await page.getByRole('button', { name: 'Save and continue' }).click();

    await expect(page).toHaveURL(/\/choose-attorneys-summary/);
    await page.getByLabel('No').check();
    await screenshot(page)
    await extractTextFromMainAndSave(page)
    await page.getByRole('button', { name: 'Continue' }).click();

    await expect(page).toHaveURL(/\/task-list/);
    await page.getByRole('link', { name: 'Choose your replacement attorneys' }).click();
    await page.getByLabel('Yes, I want replacement attorneys').check();
    await screenshot(page)
    await extractTextFromMainAndSave(page)
    await page.getByRole('button', { name: 'Save and continue' }).click();

    await expect(page).toHaveURL(/\/choose-replacement-attorneys/);
    await page.getByLabel('First names').fill('Testy');
    await page.getByLabel('Last name').fill('Testingtons');
    await page.getByLabel('Day').fill('01');
    await page.getByLabel('Month').fill('0');
    await page.getByLabel('Month').fill('01');
    await page.getByLabel('Year').fill('2000');
    await screenshot(page)
    await extractTextFromMainAndSave(page)
    await page.getByRole('button', { name: 'Save and continue' }).click();

    await expect(page).toHaveURL(/\/choose-replacement-attorneys-address/);
    await page.getByLabel('Use an address you’ve already').check();
    await screenshot(page)
    await extractTextFromMainAndSave(page)
    await page.getByRole('button', { name: 'Continue' }).click();

    await expect(page).toHaveURL(/\/choose-replacement-attorneys-address/);
    await page.getByText('1 RICHMOND PLACE').click();
    await screenshot(page)
    await extractTextFromMainAndSave(page)
    await page.getByRole('button', { name: 'Save and continue' }).click();

    await expect(page).toHaveURL(/\/choose-replacement-attorneys-summary/);
    await page.getByLabel('No').check();
    await screenshot(page)
    await extractTextFromMainAndSave(page)
    await page.getByRole('button', { name: 'Continue' }).click();

    await expect(page).toHaveURL(/\/task-list/);
    await page.getByRole('link', { name: 'Choose when your LPA can be used' }).click();

    await expect(page).toHaveURL(/\/when-can-the-lpa-be-used/);
    await page.getByLabel('Whether or not I have mental capacity to make a particular decision myself').check();
    await screenshot(page)
    await extractTextFromMainAndSave(page)
    await page.getByRole('button', { name: 'Save and continue' }).click();

    await expect(page).toHaveURL(/\/task-list/);
    await page.getByRole('link', { name: 'Add restrictions to the LPA' }).click();

    await expect(page).toHaveURL(/\/restrictions/);
    await page.getByText('Show me some examples').click();
    await page.getByLabel('Restrictions and conditions (optional)').fill('Dont sell my house');
    await screenshot(page)
    await extractTextFromMainAndSave(page)
    await page.getByRole('button', { name: 'Save and continue' }).click();

    await expect(page).toHaveURL(/\/task-list/);
    await page.getByRole('link', { name: 'Choose your certificate provider' }).click();

    await expect(page).toHaveURL(/\/what-a-certificate-provider-does/);
    await screenshot(page)
    await extractTextFromMainAndSave(page)
    await page.getByRole('link', { name: 'Continue' }).click();

    await expect(page).toHaveURL(/\/choose-your-certificate-provider/);
    await page.getByText('Examples of who could act as a professional certificate provider').click();
    await screenshot(page)
    await extractTextFromMainAndSave(page)
    await page.getByRole('link', { name: 'Continue' }).click();

    await expect(page).toHaveURL(/\/certificate-provider-details/);
    await page.getByLabel('First names').fill('Joe');
    await page.getByLabel('Last name').fill('Bloggs');
    await page.getByLabel('UK mobile number', { exact: true }).fill('07535949272');
    await screenshot(page)
    await extractTextFromMainAndSave(page)
    await page.getByRole('button', { name: 'Save and continue' }).click();

    await expect(page).toHaveURL(/\/how-do-you-know-your-certificate-provider/);
    await page.getByLabel('Personally').check();
    await screenshot(page)
    await extractTextFromMainAndSave(page)
    await page.getByRole('button', { name: 'Save and continue' }).click();

    await expect(page).toHaveURL(/\/how-long-have-you-known-certificate-provider/);
    await page.getByLabel('2 years or more').check();
    await screenshot(page)
    await extractTextFromMainAndSave(page)
    await page.getByRole('button', { name: 'Save and continue' }).click();

    await expect(page).toHaveURL(/\/how-would-certificate-provider-prefer-to-carry-out-their-role/);
    await page.getByLabel('By email').check();
    await page.getByLabel('Certificate provider’s email').fill('alex.saunders@digital.justice.gov.uk');
    await screenshot(page)
    await extractTextFromMainAndSave(page)
    await page.getByRole('button', { name: 'Save and continue' }).click();

    await expect(page).toHaveURL(/\/certificate-provider-address/);
    await page.getByLabel('Use an address you’ve already').check();
    await screenshot(page)
    await extractTextFromMainAndSave(page)
    await page.getByRole('button', { name: 'Continue' }).click();

    await expect(page).toHaveURL(/\/certificate-provider-address/);
    await page.getByLabel('1 RICHMOND PLACE').check();
    await screenshot(page)
    await extractTextFromMainAndSave(page)
    await page.getByRole('button', { name: 'Save and continue' }).click();

    await expect(page).toHaveURL(/\/task-list/);
    await page.getByRole('link', { name: 'People to notify about your LPA' }).click();

    await expect(page).toHaveURL(/\/do-you-want-to-notify-people/);
    await page.getByLabel('Yes').check();
    await screenshot(page)
    await extractTextFromMainAndSave(page)
    await page.getByRole('button', { name: 'Save and continue' }).click();

    await expect(page).toHaveURL(/\/enter-person-to-notify/);
    await page.getByLabel('First names').fill('Joette');
    await page.getByLabel('Last name').fill('Blogger');
    await screenshot(page)
    await extractTextFromMainAndSave(page)
    await page.getByRole('button', { name: 'Save and continue' }).click();

    await expect(page).toHaveURL(/\/enter-person-to-notify-address/);
    await page.getByLabel('Use an address you’ve already').check();
    await screenshot(page)
    await extractTextFromMainAndSave(page)
    await page.getByRole('button', { name: 'Continue' }).click();

    await expect(page).toHaveURL(/\/enter-person-to-notify-address/);
    await page.getByLabel('1 RICHMOND PLACE').check();
    await screenshot(page)
    await extractTextFromMainAndSave(page)
    await page.getByRole('button', { name: 'Save and continue' }).click();

    await expect(page).toHaveURL(/\/choose-people-to-notify-summary/);
    await page.getByLabel('No').check();
    await screenshot(page)
    await extractTextFromMainAndSave(page)
    await page.getByRole('button', { name: 'Save and continue' }).click();

    await expect(page).toHaveURL(/\/task-list/);
    await page.getByRole('link', { name: 'Add a correspondent' }).click();

    await expect(page).toHaveURL(/\/add-correspondent/);
    await page.getByLabel('Yes').check();
    await screenshot(page)
    await extractTextFromMainAndSave(page)
    await page.getByRole('button', { name: 'Save and continue' }).click();

    await expect(page).toHaveURL(/\/enter-correspondent-details/);
    await page.getByLabel('First names').fill('Joettey');
    await page.getByLabel('Last name').fill('Bloggers');
    await page.getByLabel('Email address').fill('a@b.com');
    await page.getByLabel('Organisation (optional)').fill('Blogger corp');
    await page.getByLabel('Phone number (optional)').fill('07535949272');
    await page.getByLabel('Yes').check();
    await screenshot(page)
    await extractTextFromMainAndSave(page)
    await page.getByRole('button', { name: 'Save and continue' }).click();

    await expect(page).toHaveURL(/\/enter-correspondent-address/);
    await page.getByLabel('Use an address you’ve already').check();
    await screenshot(page)
    await extractTextFromMainAndSave(page)
    await page.getByRole('button', { name: 'Continue' }).click();

    await expect(page).toHaveURL(/\/enter-correspondent-address/);
    await page.getByLabel('1 RICHMOND PLACE').check();
    await screenshot(page)
    await extractTextFromMainAndSave(page)
    await page.getByRole('button', { name: 'Save and continue' }).click();

    await expect(page).toHaveURL(/\/task-list/);
    await page.getByRole('link', { name: 'Check and send to your certificate provider' }).click();

    await expect(page).toHaveURL(/\/confirm-your-certificate-provider-is-not-related/);
    await page.getByLabel('I confirm that my certificate provider is not related to me or any of my attorneys').check();
    await screenshot(page)
    await extractTextFromMainAndSave(page)
    await page.getByRole('button', { name: 'Continue' }).click();

    await expect(page).toHaveURL(/\/check-your-lpa/);
    await page.getByText('Why can’t I change my LPA type?').click();
    await page.getByText('What happens if I need to make changes later?').click();
    await page.getByLabel('I’ve checked this LPA and I’m happy for OPG to share it with my certificate provider').check();
    await screenshot(page)
    await extractTextFromMainAndSave(page)
    await page.getByRole('button', { name: 'Confirm' }).click();

    await expect(page).toHaveURL(/\/lpa-details-saved/);
    await screenshot(page)
    await extractTextFromMainAndSave(page)
    await page.getByRole('link', { name: 'Return to task list' }).click();

    await expect(page).toHaveURL(/\/task-list/);
    await page.getByRole('link', { name: 'Pay for the LPA' }).click();

    await expect(page).toHaveURL(/\/about-payment/);
    await page.getByText('Who is eligible for an exemption').click();
    await page.getByText('Paying a half fee based on your income').click();
    await page.getByText('Who is eligible for a repeat application discount').click();
    await page.getByText('What you need to make a hardship application').click();
    await screenshot(page)
    await extractTextFromMainAndSave(page)
    await page.getByRole('link', { name: 'Continue' }).click();

    await expect(page).toHaveURL(/\/are-you-applying-for-fee-discount-or-exemption/);
    await page.getByLabel('Yes').check();
    await screenshot(page)
    await extractTextFromMainAndSave(page)
    await page.getByRole('button', { name: 'Save and continue' }).click({ timeout: 30000 });

    await expect(page).toHaveURL(/\/which-fee-type-are-you-applying-for/);
    await page.getByLabel('No fee (an exemption)').check();
    await screenshot(page)
    await extractTextFromMainAndSave(page)
    await page.getByRole('button', { name: 'Save and continue' }).click({ timeout: 30000 });

    await expect(page).toHaveURL(/\/evidence-required/);
    await page.getByText('Eligible means-tested benefits').click();
    await screenshot(page)
    await extractTextFromMainAndSave(page)
    await page.getByRole('link', { name: 'Continue' }).click();

    await expect(page).toHaveURL(/\/how-would-you-like-to-send-evidence/);
    await page.getByLabel('Send it by post').check();
    await screenshot(page)
    await extractTextFromMainAndSave(page)
    await page.getByRole('button', { name: 'Continue' }).click();

    await expect(page).toHaveURL(/\/send-us-your-evidence-by-post/);
    await screenshot(page)
    await extractTextFromMainAndSave(page)
    await page.getByRole('button', { name: 'Continue' }).click();

    await expect(page).toHaveURL(/\/pending-payment/);
    await screenshot(page)
    await extractTextFromMainAndSave(page)
    await page.getByRole('link', { name: 'Return to task list' }).click();

    await expect(page).toHaveURL(/\/task-list/);
    await page.getByRole('link', { name: 'Confirm your identity' }).click();

    await expect(page).toHaveURL(/\/confirm-your-identity/);
    await screenshot(page)
    await extractTextFromMainAndSave(page)
    await page.getByRole('button', { name: 'Continue' }).click();

    // mock GOL
    await page.getByRole('button', { name: 'Continue' }).click();

    await expect(page).toHaveURL(/\/identity-details/);
    await page.getByLabel('Yes').check();
    await screenshot(page)
    await extractTextFromMainAndSave(page)
    await page.getByRole('button', { name: 'Continue' }).click();

    await expect(page).toHaveURL(/\/identity-details-updated/);
    await screenshot(page)
    await extractTextFromMainAndSave(page)
    await page.getByRole('link', { name: 'Continue' }).click();

    await expect(page).toHaveURL(/\/check-your-lpa/);
    await page.getByText('Why can’t I change my LPA type?').click();
    await page.getByLabel('I’ve checked this LPA and I’m happy for OPG to share it with my certificate provider, Joe Bloggs').check();
    await screenshot(page)
    await extractTextFromMainAndSave(page)
    await page.getByRole('button', { name: 'Confirm' }).click();

    await expect(page).toHaveURL(/\/lpa-details-saved/);
    await screenshot(page)
    await extractTextFromMainAndSave(page)
    await page.getByRole('link', { name: 'Return to task list' }).click();

    await expect(page).toHaveURL(/\/task-list/);
    await page.getByRole('link', { name: 'Sign the LPA' }).click();

    await expect(page).toHaveURL(/\/how-to-sign-your-lpa/);
    await screenshot(page)
    await extractTextFromMainAndSave(page)
    await page.getByRole('link', { name: 'Start' }).click();

    await expect(page).toHaveURL(/\/your-lpa-language/);
    await page.getByLabel('Register my LPA in English').check();
    await screenshot(page)
    await extractTextFromMainAndSave(page)
    await page.getByRole('button', { name: 'Save and continue' }).click();

    await expect(page).toHaveURL(/\/read-your-lpa/);
    await screenshot(page)
    await extractTextFromMainAndSave(page)
    await page.getByRole('link', { name: 'Continue' }).click();

    await expect(page).toHaveURL(/\/your-legal-rights-and-responsibilities/);
    await screenshot(page)
    await extractTextFromMainAndSave(page)
    await page.getByRole('button', { name: 'Continue to signing page' }).click();

    await expect(page).toHaveURL(/\/sign-your-lpa/);
    await page.getByText('What happens once I’ve signed?').click();
    await page.getByText('How ticking the boxes acts as your legal signature').click();
    await page.getByLabel('I want to sign this LPA as a deed').check();
    await page.getByLabel('I want to apply to register this LPA').check();
    await screenshot(page)
    await extractTextFromMainAndSave(page)
    await page.getByRole('button', { name: 'Submit my signature' }).click();

    await expect(page).toHaveURL(/\/witnessing-your-signature/);
    await screenshot(page)
    await extractTextFromMainAndSave(page)
    await page.getByRole('button', { name: 'Continue' }).click();

    await expect(page).toHaveURL(/\/witnessing-as-certificate-provider/);
    await page.getByText('I’m having a problem with the witness code').click();
    await page.getByLabel('Enter code').fill('1234');
    await screenshot(page)
    await extractTextFromMainAndSave(page)
    await page.getByRole('button', { name: 'Continue' }).click();

    await expect(page).toHaveURL(/\/you-have-submitted-your-lpa/);
    await screenshot(page)
    await extractTextFromMainAndSave(page)
    await page.getByRole('link', { name: 'Continue' }).click();

    await expect(page).toHaveURL(/\/dashboard/);
    await screenshot(page)
    await extractTextFromMainAndSave(page)
    await page.getByRole('button', { name: 'Actions' }).click();
    await page.getByRole('link', { name: 'Check LPA progress' }).click();

    await expect(page).toHaveURL(/\/progress/);
    await screenshot(page)
    await extractTextFromMainAndSave(page)
    await page.getByRole('link', { name: 'Return to ‘Manage LPAs’' }).click();

    await expect(page).toHaveURL(/\/dashboard/);
    await page.getByRole('button', { name: 'Actions' }).click();
    await page.getByRole('link', { name: 'View LPA' }).click();

    await expect(page).toHaveURL(/\/view-lpa/);
    await screenshot(page)
    await extractTextFromMainAndSave(page)
    await page.getByRole('link', { name: 'Manage your LPAs' }).click();

    await expect(page).toHaveURL(/\/dashboard/);
    await page.getByRole('button', { name: 'Actions' }).click();
    await page.getByRole('link', { name: 'Revoke LPA' }).click();

    await expect(page).toHaveURL(/\/withdraw-this-lpa/);
    await screenshot(page)
    await extractTextFromMainAndSave(page)
    await page.getByRole('button', { name: 'Revoke this LPA' }).click();

    // error with LPA store
    // await expect(page).toHaveURL(/\/lpa-withdrawn/);
});
