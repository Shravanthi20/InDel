package com.imaginai.indel.utils

import android.content.Context
import android.os.Environment
import com.imaginai.indel.data.model.Claim
import com.itextpdf.io.font.constants.StandardFonts
import com.itextpdf.kernel.font.PdfFontFactory
import com.itextpdf.kernel.geom.PageSize
import com.itextpdf.kernel.pdf.PdfDocument
import com.itextpdf.kernel.pdf.PdfWriter
import com.itextpdf.layout.Document
import com.itextpdf.layout.borders.Border
import com.itextpdf.layout.borders.SolidBorder
import com.itextpdf.layout.element.Cell
import com.itextpdf.layout.element.Paragraph
import com.itextpdf.layout.element.Table
import com.itextpdf.layout.properties.TextAlignment
import com.itextpdf.layout.properties.VerticalAlignment
import java.io.File
import java.time.LocalDateTime
import java.time.format.DateTimeFormatter

object ClaimPdfGenerator {
    
    fun generateClaimPdf(context: Context, claim: Claim): File? {
        return try {
            // Create downloads directory
            val downloadDir = File(
                Environment.getExternalStoragePublicDirectory(Environment.DIRECTORY_DOWNLOADS),
                "InDel_Claims"
            )
            if (!downloadDir.exists()) {
                downloadDir.mkdirs()
            }
            
            // Create PDF file with timestamp
            val timestamp = LocalDateTime.now().format(DateTimeFormatter.ofPattern("yyyyMMdd_HHmmss"))
            val fileName = "Claim_${claim.claimId.replace("-", "")}_$timestamp.pdf"
            val pdfFile = File(downloadDir, fileName)
            
            // Create PDF document
            val writer = PdfWriter(pdfFile)
            val pdf = PdfDocument(writer)
            val document = Document(pdf, PageSize.A4)
            
            val font = PdfFontFactory.createFont(StandardFonts.HELVETICA)
            val fontBold = PdfFontFactory.createFont(StandardFonts.HELVETICA_BOLD)
            val fontLarge = PdfFontFactory.createFont(StandardFonts.HELVETICA_BOLD)
            
            // Set margins
            document.setMargins(20f, 20f, 20f, 20f)
            
            // Header
            val headerTable = Table(floatArrayOf(100f, 100f, 100f))
            headerTable.setWidth(555f)
            headerTable.addCell(
                Cell()
                    .add(Paragraph("InDel Insurance").setFont(fontLarge).setFontSize(18f))
                    .setBorder(Border.NO_BORDER)
                    .setPadding(0f)
            )
            headerTable.addCell(
                Cell()
                    .add(Paragraph("Claim Document").setFont(fontBold).setFontSize(12f).setTextAlignment(TextAlignment.CENTER))
                    .setBorder(Border.NO_BORDER)
                    .setPadding(0f)
            )
            headerTable.addCell(
                Cell()
                    .add(Paragraph("${LocalDateTime.now().format(DateTimeFormatter.ofPattern("dd/MM/yyyy HH:mm"))}").setFont(font).setFontSize(10f).setTextAlignment(TextAlignment.RIGHT))
                    .setBorder(Border.NO_BORDER)
                    .setPadding(0f)
            )
            document.add(headerTable)
            document.add(Paragraph("\n"))
            
            // Title
            document.add(
                Paragraph("Claim Details")
                    .setFont(fontBold)
                    .setFontSize(14f)
                    .setTextAlignment(TextAlignment.CENTER)
            )
            document.add(Paragraph("\n"))
            
            // Claim Summary Section
            document.add(Paragraph("Claim Summary").setFont(fontBold).setFontSize(12f))
            val summaryBorder = SolidBorder(0.5f)
            val summaryTable = Table(floatArrayOf(150f, 200f))
            summaryTable.setWidth(350f)
            
            // Helper function to add row
            fun addRow(label: String, value: String) {
                summaryTable.addCell(
                    Cell()
                        .add(Paragraph(label).setFont(fontBold).setFontSize(10f))
                        .setBorder(summaryBorder)
                        .setPadding(8f)
                        .setVerticalAlignment(VerticalAlignment.MIDDLE)
                )
                summaryTable.addCell(
                    Cell()
                        .add(Paragraph(value).setFont(font).setFontSize(10f))
                        .setBorder(summaryBorder)
                        .setPadding(8f)
                        .setVerticalAlignment(VerticalAlignment.MIDDLE)
                )
            }
            
            // Add summary data
            addRow("Claim ID:", claim.claimId)
            addRow("Status:", claim.status.uppercase())
            addRow("Zone:", claim.zone)
            addRow("Disruption Type:", claim.disruptionType.replace("_", " ").uppercase())
            addRow("Fraud Verdict:", claim.fraudVerdict.uppercase())
            addRow("Fraud Score:", String.format("%.2f%%", (claim.fraudScore ?: 0.0) * 100))
            addRow("Income Loss:", "₹${claim.incomeLoss}")
            addRow("Payout Amount:", "₹${claim.payoutAmount}")
            addRow("Claim Reason:", claim.claimReason ?: "N/A")
            addRow("Main Cause:", claim.mainCause ?: "N/A")
            claim.createdAt?.let {
                addRow("Claim Date:", it.take(16))
            }
            
            document.add(summaryTable)
            document.add(Paragraph("\n"))
            
            // Calculation Details
            if (!claim.calculation.isNullOrEmpty()) {
                document.add(Paragraph("Calculation Details").setFont(fontBold).setFontSize(12f))
                document.add(Paragraph(claim.calculation).setFont(font).setFontSize(10f).setTextAlignment(TextAlignment.JUSTIFIED))
                document.add(Paragraph("\n"))
            }
            
            // Contributing Factors
            if (claim.factors.isNotEmpty()) {
                document.add(Paragraph("Contributing Factors").setFont(fontBold).setFontSize(12f))
                val factorsTable = Table(floatArrayOf(250f, 100f))
                factorsTable.setWidth(350f)
                
                // Header row
                factorsTable.addCell(
                    Cell()
                        .add(Paragraph("Factor").setFont(fontBold).setFontSize(10f))
                        .setBorder(summaryBorder)
                        .setPadding(8f)
                        .setBackgroundColor(com.itextpdf.kernel.colors.ColorConstants.LIGHT_GRAY)
                )
                factorsTable.addCell(
                    Cell()
                        .add(Paragraph("Impact").setFont(fontBold).setFontSize(10f))
                        .setBorder(summaryBorder)
                        .setPadding(8f)
                        .setTextAlignment(TextAlignment.CENTER)
                        .setBackgroundColor(com.itextpdf.kernel.colors.ColorConstants.LIGHT_GRAY)
                )
                
                // Factor rows
                claim.factors.forEach { factor ->
                    factorsTable.addCell(
                        Cell()
                            .add(Paragraph(factor.name.replace("_", " ")).setFont(font).setFontSize(9f))
                            .setBorder(summaryBorder)
                            .setPadding(8f)
                    )
                    factorsTable.addCell(
                        Cell()
                            .add(Paragraph(String.format("%.2f", factor.impact)).setFont(font).setFontSize(9f))
                            .setBorder(summaryBorder)
                            .setPadding(8f)
                            .setTextAlignment(TextAlignment.CENTER)
                    )
                }
                
                document.add(factorsTable)
                document.add(Paragraph("\n"))
            }
            
            // Disruption Window
            claim.disruptionWindow?.let {
                document.add(Paragraph("Incident Window").setFont(fontBold).setFontSize(12f))
                document.add(Paragraph("Start: ${it.start}").setFont(font).setFontSize(10f))
                document.add(Paragraph("End: ${it.end}").setFont(font).setFontSize(10f))
                document.add(Paragraph("\n"))
            }
            
            // Footer
            document.add(Paragraph("\n"))
            document.add(
                Paragraph("This is an automated claim document. For disputes or clarifications, please contact support.")
                    .setFont(font)
                    .setFontSize(8f)
                    .setTextAlignment(TextAlignment.CENTER)
                    .setFontColor(com.itextpdf.kernel.colors.ColorConstants.DARK_GRAY)
            )
            
            // Close document
            document.close()
            
            pdfFile
        } catch (e: Exception) {
            e.printStackTrace()
            null
        }
    }
}
