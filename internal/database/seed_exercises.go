package database

import (
	"log"
	"time"

	"gorm.io/gorm"
	"smg-backend/internal/models"
)

func getSeedExercises() []models.Exercise {
	now := time.Now()
	return []models.Exercise{
		// ═══════════════════════════════════════════════════════════════════════
		// COURSE 1: Android Hardware Repairing Level 1
		// ═══════════════════════════════════════════════════════════════════════
		{
			ID:              "ex_c1_1",
			CourseID:        "c1",
			CourseSlug:      "android-hardware-repairing-level-1",
			ModuleID:        "m1_1",
			LessonID:        "l_c1_2",
			Title:           "Lab Exercise 1: Primary VPH_PWR Rail Diode Matrix & Cold Testing",
			Slug:            "lab-vph-pwr-diode-matrix",
			Type:            "DIODE_MEASUREMENT",
			Difficulty:      "Beginner",
			EstimatedMins:   30,
			Objective:       "Accurately locate, measure, and record the forward voltage diode drop across the primary VPH_PWR switching inductor on a live Android logic board.",
			Instructions:    "1. Disconnect all power sources and discharge any residual capacitance on the VBUS line.\n2. Set your digital multimeter to Diode Mode with continuity buzzer enabled.\n3. Firmly seat the RED probe onto the main SIM tray shielding ground frame.\n4. Place the BLACK probe on the positive terminal of the main VPH_PWR switching inductor (e.g., L1001 near PM6150).\n5. Record the exact diode voltage drop in millivolts.\n6. Repeat for all 4 corners of surrounding input decoupling ceramic capacitors.\n7. Compare readings against the master reference chart (380mV - 450mV).",
			ToolsRequired:   "Sunshine DT-17N or Fluke 17B+ Digital Multimeter, Fine needle gold probes, ESD workbench mat",
			ExpectedResults: "Diode reading must fall between 380mV and 450mV with no continuous short buzzer beep. Reverse measurement (black on GND, red on coil) should show Open Loop (OL) or > 1.2V.",
			SafetyChecklist: "• Never connect battery while multimeter is in Resistance or Diode mode\n• Ensure probes do not slip and bridge adjacent SMD capacitor pads\n• Wear ESD anti-static wrist strap connected to earth ground",
			PassingCriteria: "Documented diode drop falls within ±10% of reference value with zero grounded shorts.",
			OrderNum:        1,
			CreatedAt:       now,
		},
		{
			ID:              "ex_c1_2",
			CourseID:        "c1",
			CourseSlug:      "android-hardware-repairing-level-1",
			ModuleID:        "m1_3",
			LessonID:        "l_c1_10",
			Title:           "Lab Exercise 2: Rosin Smoke Sublimation & Voltage Injection Short Finding",
			Slug:            "lab-rosin-smoke-short-hunting",
			Type:            "PRACTICAL_LAB",
			Difficulty:      "Intermediate",
			EstimatedMins:   35,
			Objective:       "Locate and desolder an intentional shorted ceramic filter capacitor on the VPH line using pure colophony rosin vapor and current-limited DC injection.",
			Instructions:    "1. Mount the shorted practice motherboard in a PCB clamp under the microscope.\n2. Heat the rosin atomizer pen and gently frost the suspected PMIC area with thick white rosin aerosol.\n3. Configure your DC power supply to 2.8V with maximum current limited to 1.5A.\n4. Connect the negative DC alligator lead to the USB-C shield ground.\n5. Tap the positive needle probe to the VPH inductor test point for 2 to 3 seconds.\n6. Observe which capacitor frost melts transparent first under the microscope.\n7. Use hot air at 340°C with flux to lift the defective capacitor.\n8. Verify line impedance restores to > 400mV diode drop.",
			ToolsRequired:   "Electronic Rosin Atomizer, Solid Pine Rosin, 4-Channel DC Power Supply, Quick 861DW Hot Air Station, Amtech NC-559 Flux",
			ExpectedResults: "Shorted capacitor melts white frost completely within 1.5 seconds while surrounding healthy components remain frosted.",
			SafetyChecklist: "• Voltage must NOT exceed 3.5V to avoid damaging secondary PMIC input gates\n• Ensure proper bench fume extraction when vaporizing rosin flux\n• Do not apply current continuously for more than 5 seconds at a time",
			PassingCriteria: "Defective capacitor cleanly removed without PCB pad lift; circuit diode reading restored to > 380mV.",
			OrderNum:        2,
			CreatedAt:       now,
		},
		{
			ID:              "ex_c1_3",
			CourseID:        "c1",
			CourseSlug:      "android-hardware-repairing-level-1",
			ModuleID:        "m1_4",
			LessonID:        "l_c1_15",
			Title:           "Lab Exercise 3: Micro Jumper Wire 0.01mm Underfill Pad Restoration",
			Slug:            "lab-micro-jumper-pad-restoration",
			Type:            "TRACE_REPAIR",
			Difficulty:      "Advanced",
			EstimatedMins:   45,
			Objective:       "Reconstruct a torn BGA pad under an audio codec or power IC using 0.01mm insulated silver jumper wire and UV solder mask curing.",
			Instructions:    "1. Clean the missing pad location with solder wick and isopropyl alcohol.\n2. Under 25x microscope magnification, use a surgical #11 blade to gently scrape the solder mask off the buried copper via trace.\n3. Apply a microscopic dab of Amtech flux and tin the copper via with leaded solder.\n4. Weld 0.01mm enamelled jumper wire directly onto the tinned via.\n5. Wind the wire into a tight 0.3mm concentric circle matching the original BGA pad diameter.\n6. Apply a pin-point dot of green UV solder mask over the connecting trace while leaving the center coil exposed.\n7. Cure under 365nm UV lamp for 15 seconds until rock-hard.\n8. Lightly scrape the coil surface with the blade and apply a micro solder ball.",
			ToolsRequired:   "JBC C210 Soldering Station with C210-020 tip, Mechanic 0.01mm Jumper Wire, UV Solder Mask Green Oil, 365nm UV Curing Light, Microscope",
			ExpectedResults: "New pad adheres firmly to PCB, withstands gentle tweezers tug test, and measures < 0.2 ohms continuity to circuit net.",
			SafetyChecklist: "• Do not press too hard with scalpel to avoid severing multi-layer inner traces\n• Avoid UV light direct exposure to eyes",
			PassingCriteria: "Reconstructed pad stays firmly attached during heat reflow test without shifting.",
			OrderNum:        3,
			CreatedAt:       now,
		},

		// ═══════════════════════════════════════════════════════════════════════
		// COURSE 2: iPhone Motherboard & Display Masterclass
		// ═══════════════════════════════════════════════════════════════════════
		{
			ID:              "ex_c2_1",
			CourseID:        "c2",
			CourseSlug:      "iphone-display-repairing",
			ModuleID:        "m2_2",
			LessonID:        "l_c2_5",
			Title:           "Lab Exercise 4: iPhone Sandwich Motherboard Safe Separation",
			Slug:            "lab-iphone-sandwich-separation",
			Type:            "BOARD_DIAGNOSIS",
			Difficulty:      "Intermediate",
			EstimatedMins:   40,
			Objective:       "Safely split an iPhone 13 Pro / 14 Pro Max logic board sandwich using a controlled digital preheating platform without thermal blistering.",
			Instructions:    "1. Remove all logic board mounting screws and camera brackets.\n2. Secure the logic board into the silicone-lined separation platform mold.\n3. Set platform thermostat to 195°C and initiate preheat cycle.\n4. Monitor the peripheral interposer solder joints under microscope lighting.\n5. At approximately 120 seconds, observe when the test solder bead becomes molten.\n6. Using anti-static curved tweezers, gently lift the top RF board vertically without twisting.\n7. Turn off heating platform immediately and allow both halves to cool to 50°C.\n8. Inspect both upper and lower boards for torn interposer pads under 15x zoom.",
			ToolsRequired:   "MianJing / QianLi iPhone Separation Platform, High-precision tweezers, High-temperature tape, Microscope",
			ExpectedResults: "Clean separation of both boards with 0 torn pads, 0 blistered layers, and 0 solder bridges to nearby underfilled ICs.",
			SafetyChecklist: "• Never use prying force when separating; board must lift effortlessly when solder is liquidus\n• Do not heat board above 205°C to protect A-series CPU on reverse side",
			PassingCriteria: "Both boards intact with 100% of interposer pads clean and free from delamination.",
			OrderNum:        4,
			CreatedAt:       now,
		},
		{
			ID:              "ex_c2_2",
			CourseID:        "c2",
			CourseSlug:      "iphone-display-repairing",
			ModuleID:        "m2_2",
			LessonID:        "l_c2_7",
			Title:           "Lab Exercise 5: Interposer Frame Reballing with 138°C Low-Melt Paste",
			Slug:            "lab-interposer-reballing-low-temp",
			Type:            "REBALLING",
			Difficulty:      "Advanced",
			EstimatedMins:   45,
			Objective:       "Apply magnetic stencil and reball all 380+ perimeter solder pads on the iPhone interposer ring with uniform spherical solder balls.",
			Instructions:    "1. Clean the interposer rim with Goot Wick and soldering iron at 320°C until all pads are flat.\n2. Clean flux residue with 99% anhydrous isopropyl alcohol.\n3. Position the dedicated magnetic reballing stencil over the bottom motherboard.\n4. Scrape dry 138°C bismuth-tin solder paste evenly across all stencil apertures using a squeegee.\n5. Blot excess moisture with lint-free wipes.\n6. Apply hot air at 260°C 40 airflow from 25mm distance until paste coalesces into shiny solder balls.\n7. Allow stencil to cool, remove magnetic clamp, and inspect ball uniformity under microscope.",
			ToolsRequired:   "Amaoe / QianLi Magnetic Reballing Platform, 138°C Low-Temperature Solder Paste, Quick 861DW, Goot Wick, Isopropyl Alcohol",
			ExpectedResults: "Every single interposer pad possesses an identical 0.25mm height solder sphere with zero bridging.",
			SafetyChecklist: "• Clean stencil apertures thoroughly before application\n• Solder paste must be partially dried with a paper towel to prevent bubbling during reflow",
			PassingCriteria: "100% of interposer balls formed with zero bridges and uniform coplanar height.",
			OrderNum:        5,
			CreatedAt:       now,
		},
		{
			ID:              "ex_c2_3",
			CourseID:        "c2",
			CourseSlug:      "iphone-display-repairing",
			ModuleID:        "m2_3",
			LessonID:        "l_c2_10",
			Title:           "Lab Exercise 6: iPhone Display IC Swap & True Tone Serialization",
			Slug:            "lab-display-ic-swap-true-tone",
			Type:            "COMPONENT_SWAP",
			Difficulty:      "Advanced",
			EstimatedMins:   40,
			Objective:       "Transfer the serialized touch EEPROM chip from an original broken iPhone screen to an aftermarket OLED assembly to eliminate the 'Unknown Part' warning.",
			Instructions:    "1. Peel back the protective graphite film behind the original display flex cable.\n2. Shield nearby flex connections with Kapton tape.\n3. Apply 280°C hot air at 30 airflow to soften the epoxy around the 8-ball EEPROM IC.\n4. Lift the IC with fine tweezers.\n5. Reball the EEPROM using a 0.10mm mini-stencil and 183°C paste.\n6. Align and solder the IC onto the replacement screen flex at 300°C.\n7. Connect JCID V1SE programmer and restore True Tone display calibration data.",
			ToolsRequired:   "JCID V1SE / QianLi iCopy Programmer, Hot Air Station, Micro Tweezers, Amtech Flux, Kapton Heat Tape",
			ExpectedResults: "Display boots without 'Non-genuine Display' warning in iOS Settings; True Tone toggle restored in Control Center.",
			SafetyChecklist: "• Do not overheat display flex; heat limit is 300°C for max 20 seconds\n• Confirm Pin 1 dot orientation before soldering",
			PassingCriteria: "Device settings show display as genuine Apple component; True Tone sensor functional.",
			OrderNum:        6,
			CreatedAt:       now,
		},

		// ═══════════════════════════════════════════════════════════════════════
		// COURSE 3: EMMC & UFS Chip-Off Data Recovery
		// ═══════════════════════════════════════════════════════════════════════
		{
			ID:              "ex_c3_1",
			CourseID:        "c3",
			CourseSlug:      "emmc-ufs-programming",
			ModuleID:        "m3_3",
			LessonID:        "l_c3_12",
			Title:           "Lab Exercise 7: In-System Programming (ISP) Micro-Wiring & Register Read",
			Slug:            "lab-isp-wiring-register-read",
			Type:            "PRACTICAL_LAB",
			Difficulty:      "Advanced",
			EstimatedMins:   50,
			Objective:       "Solder microscopic ISP jumper wires to VCC, VCCQ, CLK, CMD, and DAT0 testpoints on a dead Android board to read memory partitions without desoldering the chip.",
			Instructions:    "1. Locate the ISP pinout diagram in Borneo Schematics for the target device (e.g., Redmi Note 9).\n2. Scrape the solder mask off the CLK resistor and DAT0 capacitor under 30x zoom.\n3. Solder 0.08mm insulated copper wires from EasyJtag Plus ISP adapter to the motherboard testpoints.\n4. Keep wire length strictly under 5cm to prevent clock signal reflections.\n5. Connect EasyJtag Plus Box to PC via USB 3.0.\n6. Launch EasyJtag Classic software. Set VCC to 2.8V, VCCQ to 1.8V, and Bus Width to 1-Bit.\n7. Click 'Detect eMMC / UFS'.\n8. Verify EXT_CSD register data, CID, and partition table (ROM1, ROM2, ROM3).",
			ToolsRequired:   "EasyJtag Plus / UFI Box, 0.08mm Enamelled Wire, JBC Fine Tip Iron, Borneo Schematics, Microscope",
			ExpectedResults: "Software detects memory chip successfully, displays manufacturer ID (Samsung / Micron / SK Hynix), and lists healthy partition tables.",
			SafetyChecklist: "• Ensure wire length is < 5cm; long wires cause packet corruptions\n• Verify VCC and VCCQ voltage levels before clicking Detect to avoid blowing memory controller",
			PassingCriteria: "Full EXT_CSD dump read and verified with zero CRC transmission errors.",
			OrderNum:        7,
			CreatedAt:       now,
		},
		{
			ID:              "ex_c3_2",
			CourseID:        "c3",
			CourseSlug:      "emmc-ufs-programming",
			ModuleID:        "m3_4",
			LessonID:        "l_c3_16",
			Title:           "Lab Exercise 8: Qualcomm 9008 Sahara Firehose Dead Boot Recovery",
			Slug:            "lab-qualcomm-9008-dead-boot-recovery",
			Type:            "BOARD_DIAGNOSIS",
			Difficulty:      "Advanced",
			EstimatedMins:   45,
			Objective:       "Force a Qualcomm dead-boot device into Emergency Download Mode (EDL 9008) via hardware testpoints and flash raw partition XML files.",
			Instructions:    "1. Remove battery and disconnect camera cables.\n2. Locate the two EDL gold testpoints on the PCB layout.\n3. Short both testpoints to GND using curved metal tweezers.\n4. Insert USB-C cable connected to computer.\n5. Open Windows Device Manager and verify 'Qualcomm HS-USB QDLoader 9008' port enumeration.\n6. Open MiFlash / QFIL programmer.\n7. Load the signed Firehose programmer (prog_emmc_firehose_xxxx.mbn).\n8. Load rawprogram0.xml and patch0.xml.\n9. Flash boot, recovery, and persist partitions to restore device bootloader.",
			ToolsRequired:   "EDL Testpoint Tweezers, USB-C Data Cable, QFIL / MiFlash Tool, Qualcomm EDL Drivers",
			ExpectedResults: "Qualcomm 9008 port recognized immediately without error code 10; firmware partition flashing completes with SUCCESS status.",
			SafetyChecklist: "• Remove tweezers short immediately after USB insertion chime\n• Backup EFS (modemst1, modemst2, fsg) before flashing to safeguard IMEI",
			PassingCriteria: "Device successfully boots past Qualcomm emergency mode into Fastboot / Android OS.",
			OrderNum:        8,
			CreatedAt:       now,
		},

		// ═══════════════════════════════════════════════════════════════════════
		// COURSE 4: Borneo & ZXW Circuit Schematics Deep-Dive
		// ═══════════════════════════════════════════════════════════════════════
		{
			ID:              "ex_c4_1",
			CourseID:        "c4",
			CourseSlug:      "borneo-zxw-schematic-reading",
			ModuleID:        "m4_2",
			LessonID:        "l_c4_6",
			Title:           "Lab Exercise 9: Tracing Open-Circuit Lines in Borneo PCB Bitmap",
			Slug:            "lab-borneo-bitmap-open-line-tracing",
			Type:            "BOARD_DIAGNOSIS",
			Difficulty:      "Intermediate",
			EstimatedMins:   35,
			Objective:       "Identify an open circuit on a broken display MIPI line using Borneo Schematics interactive netlines and measure impedance verification.",
			Instructions:    "1. Open Borneo Schematics on PC and select target model (e.g., Xiaomi Redmi Note 10).\n2. Navigate to 'Hardware Solutions' -> 'LCD & Touch Light'.\n3. Click on the display FPC connector pin corresponding to MIPI_DSI0_CLK_N.\n4. Follow the highlighted magenta netline to its series filter coil and terminating CPU pin.\n5. Set multimeter to diode mode on the physical board.\n6. Measure both sides of the zero-ohm series EMI filter resistor.\n7. Locate the open circuit between the filter resistor and connector pin.\n8. Solder an insulated 0.02mm jumper wire to bridge the broken internal trace.",
			ToolsRequired:   "Borneo Schematics VIP License, Multimeter, Microscope, 0.02mm Jumper Wire, Soldering Station",
			ExpectedResults: "Diode mode on connector pin rises from 0L (Open Line) back to the factory reference 440mV ± 20mV.",
			SafetyChecklist: "• Do not touch live probe to adjacent ground pin on micro-pitch 0.35mm FPC connectors",
			PassingCriteria: "Display FPC pin shows identical diode drop to complementary differential pair.",
			OrderNum:        9,
			CreatedAt:       now,
		},

		// ═══════════════════════════════════════════════════════════════════════
		// COURSE 5: Qualcomm & MediaTek CPU Reballing Pro
		// ═══════════════════════════════════════════════════════════════════════
		{
			ID:              "ex_c5_1",
			CourseID:        "c5",
			CourseSlug:      "pmic-power-section-diagnosis",
			ModuleID:        "m5_2",
			LessonID:        "l_c5_6",
			Title:           "Lab Exercise 10: Double-Decker CPU & RAM Separation",
			Slug:            "lab-double-decker-cpu-ram-separation",
			Type:            "REBALLING",
			Difficulty:      "Master",
			EstimatedMins:   60,
			Objective:       "Separate the top LPDDR5 RAM IC from the bottom Snapdragon 888 Application Processor without pulling internal inter-die micro-balls.",
			Instructions:    "1. Mount board into high-mass CNC fixture with bottom heat sink dissipation.\n2. Clean black edge glue around the perimeter at 230°C with ultra-fine curved scraper.\n3. Increase hot air station to 345°C with 45 airflow.\n4. Direct nozzle at the top RAM surface using continuous 15mm circular pattern for 35 seconds.\n5. Gently insert 0.04mm QianLi pry blade between the RAM and CPU layer.\n6. When solder reaches liquidus, lift the RAM chip horizontally in one smooth movement.\n7. Immediately withdraw heat to prevent bottom CPU balls from expanding and blowing out.\n8. Inspect bottom CPU silicon die for blistering or micro-cracks under microscope.",
			ToolsRequired:   "Quick 861DW Station, QianLi 008 Thin Prying Blade, BGA Fixture, Amtech NC-559 Flux, Stereo Microscope",
			ExpectedResults: "Top RAM chip cleanly separated with 100% of CPU top pads intact and zero CPU bottom solder extrusion.",
			SafetyChecklist: "• Never twist or lever blade upward; slide flat horizontally\n• Exceeding 45 seconds of direct heat will destroy the silicon bonding wires",
			PassingCriteria: "Application processor pads 100% intact; chip re-recognizes over USB EDL 9008 after clean reflow.",
			OrderNum:        10,
			CreatedAt:       now,
		},
	}
}

// SeedExercises saves all course-wide practical lab exercises to MySQL
func SeedExercises(db *gorm.DB) {
	exercises := getSeedExercises()
	for _, ex := range exercises {
		var existing models.Exercise
		if err := db.Where("id = ?", ex.ID).First(&existing).Error; err != nil {
			if err := db.Create(&ex).Error; err != nil {
				log.Printf("⚠️ Failed to create exercise %s: %v", ex.ID, err)
			}
		} else {
			existing.Title = ex.Title
			existing.Slug = ex.Slug
			existing.Type = ex.Type
			existing.Difficulty = ex.Difficulty
			existing.EstimatedMins = ex.EstimatedMins
			existing.Objective = ex.Objective
			existing.Instructions = ex.Instructions
			existing.ToolsRequired = ex.ToolsRequired
			existing.ExpectedResults = ex.ExpectedResults
			existing.SafetyChecklist = ex.SafetyChecklist
			existing.PassingCriteria = ex.PassingCriteria
			existing.OrderNum = ex.OrderNum
			db.Save(&existing)
		}
	}
	log.Printf("✅ Verified / Seeded %d comprehensive practical lab exercises across courses!", len(exercises))
}
