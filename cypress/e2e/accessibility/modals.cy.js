describe('Modals', () => {
    describe('Data loss warning', () => {
        describe('Return to task list', () => {
            it('locks focus to modal', () => {
                cy.visit('/fixtures?redirect=/choose-attorneys-guidance&progress=provideYourDetails');
                cy.contains('a', 'Continue').click()

                cy.get('#f-first-names').invoke('val', 'John');
                cy.contains('a', 'Return to task list').click()

                cy.get('#data-loss-dialog').should('be.visible')

                cy.focused().should('contain', 'Back to page')

                cy.realPress("Tab")
                cy.focused().should('contain', 'Continue without saving')

                cy.realPress("Tab")
                cy.focused().should('contain', 'Back to page')

                cy.contains('button', 'Back to page').click()

                cy.get('#data-loss-dialog').should('not.be.visible')

                cy.focused().should('contain', 'Return to task list')

                cy.contains('a', 'Return to task list').click()

                cy.get('#data-loss-dialog').should('be.visible')

                cy.focused().should('contain', 'Back to page')

                cy.realPress(["Shift", "Tab"])
                cy.focused().should('contain', 'Continue without saving')

                cy.realPress(["Shift", "Tab"])
                cy.focused().should('contain', 'Back to page')

                cy.realType('{esc}')

                cy.get('#data-loss-dialog').should('not.be.visible')

                cy.focused().should('contain', 'Return to task list')
            })
        })

        describe('Change language', () => {
            it('locks focus to modal', () => {
                cy.visit('/fixtures?redirect=/choose-attorneys-guidance&progress=provideYourDetails');
                cy.contains('a', 'Continue').click()

                cy.get('#f-first-names').invoke('val', 'John');
                cy.contains('a', 'Cymraeg').click()

                cy.get('#language-dialog').should('be.visible')

                cy.focused().should('contain', 'Back to page')

                cy.realPress("Tab")
                cy.focused().should('contain', 'Continue without saving')

                cy.realPress("Tab")
                cy.focused().should('contain', 'Back to page')

                cy.contains('#language-dialog button', 'Back to page').click()

                cy.get('#data-loss-dialog').should('not.be.visible')

                cy.focused().should('contain', 'Cymraeg')

                cy.contains('a', 'Cymraeg').click()

                cy.get('#language-dialog').should('be.visible')

                cy.focused().should('contain', 'Back to page')

                cy.realPress(["Shift", "Tab"])
                cy.focused().should('contain', 'Continue without saving')

                cy.realPress(["Shift", "Tab"])
                cy.focused().should('contain', 'Back to page')

                cy.realType('{esc}')

                cy.get('#language-dialog').should('not.be.visible')

                cy.focused().should('contain', 'Cymraeg')
            })
        })
    })

    describe('File upload', () => {
        it('locks focus to modal', () => {
            cy.visit('/fixtures?redirect=/about-payment&progress=checkAndSendToYourCertificateProvider');
            cy.contains('a', 'Continue').click();
            cy.get('input[name="yes-no"]').check('yes', { force: true });
            cy.contains('button', 'Save and continue').click();
            cy.get('input[name="fee-type"]').check('HalfFee', { force: true });
            cy.contains('button', 'Save and continue').click();
            cy.get('h1').should('contain', 'Evidence required to pay a half fee');
            cy.contains('a', 'Continue').click();
            cy.get('input[name="selected"]').check('upload', { force: true });
            cy.contains('button', 'Continue').click();

            cy.get('input[type="file"]').attachFile(['dummy.pdf', 'dummy.png']);
            cy.contains('button', 'Upload files').click()

            cy.get('#file-upload-dialog').should('be.visible')

            cy.focused().should('contain', 'Cancel upload')

            cy.realPress("Tab")

            cy.focused().should('contain', 'Cancel upload')

            cy.realPress(["Shift", "Tab"])

            cy.focused().should('contain', 'Cancel upload')

            cy.realType('{esc}')

            cy.get('#file-upload-dialog').should('not.be.visible')

            // trigger not focused as cancelling SSE connection reloads page
        })
    })
})
