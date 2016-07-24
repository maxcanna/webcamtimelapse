package net.luxteam.webcamtimelapse;

import java.awt.AlphaComposite;
import java.awt.Color;
import java.awt.Font;
import java.awt.FontMetrics;
import java.awt.Graphics2D;
import java.awt.geom.Rectangle2D;
import java.awt.image.BufferedImage;
import java.io.File;
import java.io.IOException;
import java.net.MalformedURLException;
import java.net.URL;
import java.util.Date;
import java.util.Timer;
import java.util.TimerTask;

import javax.imageio.ImageIO;

import org.eclipse.swt.widgets.Display;
import org.eclipse.swt.widgets.Shell;
import org.eclipse.swt.widgets.Group;
import org.eclipse.swt.SWT;
import org.eclipse.swt.widgets.Event;
import org.eclipse.swt.widgets.FileDialog;
import org.eclipse.swt.widgets.Label;
import org.eclipse.swt.widgets.Listener;
import org.eclipse.swt.widgets.MessageBox;
import org.eclipse.swt.widgets.Text;
import org.eclipse.swt.widgets.Spinner;
import org.eclipse.swt.widgets.Button;
import org.eclipse.swt.widgets.Scale;
import org.eclipse.swt.widgets.ProgressBar;
import org.eclipse.swt.events.MouseAdapter;
import org.eclipse.swt.events.MouseEvent;

import ch.randelshofer.media.quicktime.QuickTimeOutputStream;
import ch.randelshofer.media.quicktime.QuickTimeOutputStream.VideoFormat;
import org.eclipse.swt.events.FocusAdapter;
import org.eclipse.swt.events.FocusEvent;
import org.eclipse.swt.events.ModifyListener;
import org.eclipse.swt.events.ModifyEvent;
import org.eclipse.swt.graphics.Point;

public class MainWindow {

	protected static Display display;
	protected static Shell shlWebcamTimelapse;
	
	private static Text textURL;
	private static Text textFile;
	private static Spinner spinnerInterval;
	private static Spinner spinnerFrames;
	private static Spinner spinnerFPS;
	private static Label labelWidthValue;
	private static Label labelHeightValue;
	private static Scale scaleQuality;
	private static Label labelVideoDurationValue;
	private static Label labelCaptureDurationValue;
	private static Label labelStatusValue;
	private static Button buttonStartStop;
	private static ProgressBar progressBar;

	private static BufferedImage img = null;
	private static QuickTimeOutputStream out = null;
	private static File file = null;
	private static VideoFormat format = QuickTimeOutputStream.VideoFormat.JPG;
	private static TimerTask timerTask = null;
	private static Timer timer = null;
	private static int frameCount = 0;

	//Params
	private static String url = null;
	private static String filename = null;
	private static Integer interval = null;
	private static int frames = 0;
	private static float quality = 1f;
	private static int fps = 30;
	private static int h = 0;
	private static int w = 0;
	private static boolean stop = true;
	
	/**
	 * Launch the application.
	 * @param args
	 */
	public static void main(String[] args) {
		try {
			MainWindow window = new MainWindow();
			window.open();
		} catch (Exception e) {
			e.printStackTrace();
		}
	}

	/**
	 * Open the window.
	 */
	public void open() {
		display = Display.getDefault();
		createContents();
		shlWebcamTimelapse.open();
		shlWebcamTimelapse.layout();
		while (!shlWebcamTimelapse.isDisposed()) {
			if (!display.readAndDispatch()) {
				display.sleep();
			}
		}
	}

	/**
	 * Create contents of the window.
	 */
	protected void createContents() {
		shlWebcamTimelapse = new Shell(SWT.DIALOG_TRIM | SWT.MIN);
		shlWebcamTimelapse.setSize(new Point(550, 270));
		shlWebcamTimelapse.setMinimumSize(new Point(550, 270));
		shlWebcamTimelapse.setSize(550, 270);
		shlWebcamTimelapse.setText("WebCamTimeLapse");
		
		Group groupInput = new Group(shlWebcamTimelapse, SWT.NONE);
		groupInput.setText("Input");
		groupInput.setBounds(10, 10, 511, 79);
		
		Label labelURL = new Label(groupInput, SWT.NONE);
		labelURL.setText("URL");
		labelURL.setBounds(10, 20, 47, 15);
		
		textURL = new Text(groupInput, SWT.BORDER);
		textURL.addFocusListener(new FocusAdapter() {
			@Override
			public void focusLost(FocusEvent arg0) {
				url = textURL.getText();
				if(url.contains("jpg")||url.contains("jpeg")) filename = url.substring(url.lastIndexOf("/") + 1, url.length()).replace(".jpg", "").replace(".jpeg", "")+".mov";
			    else filename = "Output_"+System.currentTimeMillis()+".mov";
			}
		});
		textURL.setToolTipText("Image URL");
		textURL.setBounds(63, 17, 442, 21);
		
		spinnerInterval = new Spinner(groupInput, SWT.BORDER);
		spinnerInterval.setMaximum(3600);
		spinnerInterval.setTextLimit(3600);
		spinnerInterval.setMinimum(10);
		spinnerInterval.setSelection(30);
		spinnerInterval.addModifyListener(new ModifyListener() {
			public void modifyText(ModifyEvent arg0) {
				try {
					if(Integer.parseInt(spinnerFrames.getText()) == 0)labelCaptureDurationValue.setText("-");
					else labelCaptureDurationValue.setText(Integer.parseInt(spinnerInterval.getText())*Integer.parseInt(spinnerFrames.getText())+"\"");
				} catch (NumberFormatException e) {
					spinnerFrames.setSelection(50);
				}
			}
		});
		spinnerInterval.setToolTipText("Time between capture");
		spinnerInterval.setBounds(63, 44, 47, 22);
		
		Label labelInterval = new Label(groupInput, SWT.NONE);
		labelInterval.setText("Interval");
		labelInterval.setBounds(10, 47, 47, 15);
		
		Label labelFrames = new Label(groupInput, SWT.NONE);
		labelFrames.setText("Frames");
		labelFrames.setBounds(116, 47, 38, 15);
		
		spinnerFrames = new Spinner(groupInput, SWT.BORDER);
		spinnerFrames.setSelection(50);
		spinnerFrames.setToolTipText("Number of frame acquired. Set 0 for unlimited");
		spinnerFrames.setBounds(160, 44, 55, 22);		
		spinnerFrames.addModifyListener(new ModifyListener() {
			public void modifyText(ModifyEvent arg0) {
				try {
					if(Integer.parseInt(spinnerFrames.getText()) == 0){
						labelVideoDurationValue.setText("-");
						labelCaptureDurationValue.setText("-");
					}
					else{
						labelVideoDurationValue.setText(Float.parseFloat(spinnerFrames.getText())/Float.parseFloat(spinnerFPS.getText())+"\"");
						labelCaptureDurationValue.setText(Integer.parseInt(spinnerInterval.getText())*Integer.parseInt(spinnerFrames.getText())+"\"");
					}
				} catch (NumberFormatException e) {
					spinnerFrames.setSelection(50);
				}
			}
		});
		
		Group groupOutput = new Group(shlWebcamTimelapse, SWT.NONE);
		groupOutput.setText("Output");
		groupOutput.setBounds(10, 92, 514, 90);
		
		spinnerFPS = new Spinner(groupOutput, SWT.BORDER);
		spinnerFPS.setSelection(25);
		spinnerFPS.setToolTipText("Frame per second");
		spinnerFPS.setBounds(299, 52, 47, 22);
		spinnerFPS.addModifyListener(new ModifyListener() {
			public void modifyText(ModifyEvent arg0) {
				try {
					if(Integer.parseInt(spinnerFrames.getText()) == 0) labelVideoDurationValue.setText("-");
					else labelVideoDurationValue.setText(Float.parseFloat(spinnerFrames.getText())/Float.parseFloat(spinnerFPS.getText())+"\"");
				} catch (NumberFormatException e) {
					spinnerFrames.setSelection(50);
				}
			}
		});
		
		Label labelCaptureDuration = new Label(groupInput, SWT.NONE);
		labelCaptureDuration.setText("Job Duration");
		labelCaptureDuration.setBounds(221, 47, 67, 15);
		
		labelCaptureDurationValue = new Label(groupInput, SWT.NONE);
		labelCaptureDurationValue.setText("120\"");
		labelCaptureDurationValue.setBounds(294, 47, 38, 15);
		
		Label labelWidth = new Label(groupInput, SWT.NONE);
		labelWidth.setBounds(338, 47, 40, 15);
		labelWidth.setText("Width");
		
		labelWidthValue = new Label(groupInput, SWT.NONE);
		labelWidthValue.setBounds(384, 47, 35, 15);
		labelWidthValue.setToolTipText("Frame per second");
		
		Label labelHeight = new Label(groupInput, SWT.NONE);
		labelHeight.setText("Height");
		labelHeight.setBounds(426, 47, 38, 15);
		
		labelHeightValue = new Label(groupInput, SWT.NONE);
		labelHeightValue.setToolTipText("");
		labelHeightValue.setBounds(470, 47, 35, 15);
		
		Label labelFile = new Label(groupOutput, SWT.NONE);
		labelFile.setText("File");
		labelFile.setBounds(10, 20, 30, 15);
		
		textFile = new Text(groupOutput, SWT.BORDER);
		textFile.setToolTipText("Output filename");
		textFile.setBounds(63, 17, 409, 21);
		
		Button buttonBrowse = new Button(groupOutput, SWT.NONE);
		buttonBrowse.setText("...");
		buttonBrowse.setBounds(478, 15, 30, 25);
		buttonBrowse.addMouseListener(new MouseAdapter() {
			@Override
			public void mouseUp(MouseEvent arg0) {
				FileDialog dialog = new FileDialog(shlWebcamTimelapse, SWT.SAVE);
			    dialog.setFilterNames(new String[] { "QuickTime File" });
			    dialog.setFilterExtensions(new String[] { "*.mov" });
			    dialog.setFilterPath(new File(".").getAbsolutePath()); // Windows path

			    dialog.setFileName(filename);
			    textFile.setText(dialog.open());
			}
		});
		
		Label labelFPS = new Label(groupOutput, SWT.NONE);
		labelFPS.setText("FPS");
		labelFPS.setBounds(263, 55, 30, 15);
		
		final Label labelQualityValue = new Label(groupOutput, SWT.NONE);
		labelQualityValue.setText("9");
		labelQualityValue.setBounds(231, 55, 30, 15);
		
		scaleQuality = new Scale(groupOutput, SWT.NONE);
		scaleQuality.addListener(SWT.Selection, new Listener() {
			public void handleEvent(Event event) {
				labelQualityValue.setText(""+scaleQuality.getSelection());
			}
		});

		scaleQuality.setToolTipText("Quelity level. Higher is better but produces bigger file");
		scaleQuality.setPageIncrement(1);
		scaleQuality.setMaximum(10);
		scaleQuality.setMinimum(1);
		scaleQuality.setSelection(9);
		scaleQuality.setBounds(63, 44, 162, 42);
		
		Label labelQuality = new Label(groupOutput, SWT.NONE);
		labelQuality.setText("Quality");
		labelQuality.setBounds(10, 55, 47, 15);
		
		Label lblVideoDuration = new Label(groupOutput, SWT.NONE);
		lblVideoDuration.setText("Video Duration");
		lblVideoDuration.setBounds(360, 55, 79, 15);
		
		labelVideoDurationValue = new Label(groupOutput, SWT.NONE);
		labelVideoDurationValue.setText("2\"");
		labelVideoDurationValue.setBounds(443, 55, 61, 15);
		
		buttonStartStop = new Button(shlWebcamTimelapse, SWT.NONE);
		buttonStartStop.addMouseListener(new MouseAdapter() {
			@Override
			public void mouseUp(MouseEvent arg0) {
				if(timerTask == null) doJob();
				else stopJob();
				buttonStartStop.setText(timerTask == null ? "Start": "Stop");
			}
		});
		buttonStartStop.setText("Start");
		buttonStartStop.setBounds(10, 188, 36, 25);

		progressBar = new ProgressBar(shlWebcamTimelapse, SWT.SMOOTH);
		progressBar.setVisible(false);
		progressBar.setBounds(52, 188, 472, 25);
		
		Label labelStatus = new Label(shlWebcamTimelapse, SWT.NONE);
		labelStatus.setText("Status");
		labelStatus.setBounds(12, 219, 32, 15);
		
		labelStatusValue = new Label(shlWebcamTimelapse, SWT.NONE);
		labelStatusValue.setBounds(52, 219, 469, 15);

	}
	
	private static void stopJob(){
		stop = true;
		frameCount = 0;
		display.asyncExec(new Runnable(){
			@Override
			public void run() {
				buttonStartStop.setText(stop ? "Start": "Stop");
				
				if (out != null) {
					try {
						out.close();
					}
					catch (IOException e) {
						labelStatusValue.setText(new Date().toString()+" Error closing file!");
					}
				}
				labelStatusValue.setText(new Date().toString()+" Done!");				
			}
		});
		
		timer = null;
		timerTask.cancel();
		timerTask = null;
	}
	
	private static void doJob(){
		stop = false;
		buttonStartStop.setText(stop ? "Start": "Stop");
		progressBar.setVisible(false);
		try {
			url = textURL.getText();
			//if(!url.contains(".jpg") && url.contains(".jpeg")) throw new Exception("Not a jpg image in URL!");
			//Testo validita' URL e che all'indirizzo ci sia un'img valida
			img = ImageIO.read(new URL(url));
			if(img == null) throw new Exception("Not a jpg image in URL!");
		} catch (Exception e) {
			MessageBox messagebox = new MessageBox(shlWebcamTimelapse,SWT.ICON_ERROR);
			messagebox.setMessage("Invalid URL. Not a jpg image in URL");
			messagebox.setText("Error");
			messagebox.open();
			return;
		}
		
		try{
			filename = textFile.getText();
			if(filename.equals("")){
				if(url.contains("jpg")||url.contains("jpeg")) filename = url.substring(url.lastIndexOf("/") + 1, url.length()).replace(".jpg", "").replace(".jpeg", "")+".mov";
			    else filename = "Output_"+System.currentTimeMillis()+".mov";
				filename = new File(".").getAbsolutePath() + "\\" + filename;
				textFile.setText(filename);
			}
			//Se il file esiste lo rinomino
			if(new File(filename).exists()){
				MessageBox messagebox = new MessageBox(shlWebcamTimelapse,SWT.ICON_WARNING);
				messagebox.setMessage("File already exists.\nRenamed to "+filename.replace(".mov", "")+"_old.mov");
				messagebox.setText("Warning");
				messagebox.open();
				new File(filename).renameTo(new File(filename+"_old.mov"));
			}
			//Provo a vedere se posso scrivere
			new File(filename).canWrite();
		} catch (Exception e1) {	
			//Se il file esiste lo rinomino
			if(new File(filename).exists()){
				MessageBox messagebox = new MessageBox(shlWebcamTimelapse,SWT.ICON_WARNING);
				messagebox.setMessage("File already exists.\nRenamed to "+filename.replace(".mov", "")+"_old.mov");
				messagebox.setText("Warning");
				messagebox.open();
				new File(filename).renameTo(new File(filename+"_old.mov"));
			}
		}
		
		try{
			interval = Integer.parseInt(spinnerInterval.getText());
		} catch(Exception e) {
			interval = 120;
			spinnerInterval.setSelection(interval);
			MessageBox messagebox = new MessageBox(shlWebcamTimelapse,SWT.ICON_WARNING);
			messagebox.setMessage("Invalid frames interval. Using "+interval);
			messagebox.setText("Warning");
			messagebox.open();
		}
		
		try{
			frames = Integer.parseInt(spinnerFrames.getText());
		} catch(Exception e) {
			frames = 0;
			MessageBox messagebox = new MessageBox(shlWebcamTimelapse,SWT.ICON_WARNING);
			messagebox.setMessage("Invalid frames count. Using infinite.");
			messagebox.setText("Warning");
			messagebox.open();
		}
		
		try{
			fps = Integer.parseInt(spinnerFPS.getText());
		} catch(Exception e){
			fps = 15;
			spinnerFPS.setSelection(fps);
			MessageBox messagebox = new MessageBox(shlWebcamTimelapse,SWT.ICON_WARNING);
			messagebox.setMessage("Invalid FPS. Using "+fps);
			messagebox.setText("Warning");
			messagebox.open();
		}
		
		try{
			if(img == null) img = ImageIO.read(new URL(url));
			w = img.getWidth();
			h = img.getHeight();
			labelWidthValue.setText(""+w);
			labelHeightValue.setText(""+h);
		}
		catch (MalformedURLException e1) { }
		catch (IOException e1) { }
		
		quality = ((float)scaleQuality.getSelection())/10;

		try {
			file = new File(filename);
			out = new QuickTimeOutputStream(file, format);
			out.setVideoCompressionQuality(quality);
			out.setTimeScale(fps);
			out.setVideoDimension(w,h);
		}
		catch (IOException e) { }

		timerTask = new TimerTask() {
			public void run() {
				if(!stop)addFrame();
			}
		};

		synchronized (timerTask) {
			timer = new Timer();
			timer.scheduleAtFixedRate(timerTask, 0, interval*1000);
			labelStatusValue.setText(new Date().toString()+" Capture job started...");
			progressBar.setVisible(frames != 0);
		}
	}
	
	private static void addFrame(){
		try{
			img = watermark(ImageIO.read(new URL(url)),"WebCamTimeLapse");
			out.writeFrame(img, 1);
			frameCount++;
			display.asyncExec(new Runnable(){
				@Override
				public void run() {
					labelStatusValue.setText(new Date().toString()+" Frame "+frameCount+" added.");
					if(frames > 0){
						labelStatusValue.setText(labelStatusValue.getText()+" "+(frames*interval/60)+" minutes remaining");
						progressBar.setSelection((int)((float)frameCount/(float)frames*100.0f));
					}else{
						labelVideoDurationValue.setText(""+(float)frameCount/(float)fps);
						labelCaptureDurationValue.setText(frameCount*interval+"\"");
					}
				}
			});
			synchronized (timerTask) {					
				if (frameCount == frames) stopJob();
					//timerTask.notify();
			}
		}
		catch (MalformedURLException e) {}
		catch (IOException e) {
			labelStatusValue.setText(new Date().toString()+" Error adding frame.");
		}
	}
	
	private static BufferedImage watermark(BufferedImage image,String text) {
		try {
			AlphaComposite alpha = AlphaComposite.getInstance(AlphaComposite.SRC_OVER,0.5f);  
			Graphics2D g = image.createGraphics();
			Font font = new Font("Arial", Font.BOLD, 20);
			FontMetrics fontMetrics = g.getFontMetrics();  
			Rectangle2D rect = fontMetrics.getStringBounds(text, g);  
			/*int centerX = (w - (int) rect.getWidth()) / 2;  
			int centerY = (h - (int) rect.getHeight()) / 2;  			*/
			g.setColor(Color.RED);
			g.setFont(font);
			g.setComposite(alpha);  
			g.drawString(text, 0, (int)rect.getHeight()*2); 
			g.dispose();
		}
		catch (Exception e){ }
		return image;
	}
}
