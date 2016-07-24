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

import ch.randelshofer.media.quicktime.QuickTimeOutputStream;
import ch.randelshofer.media.quicktime.QuickTimeOutputStream.VideoFormat;

public class WebCamTimeLapse implements Runnable {

	private static BufferedImage img = null;

	private static QuickTimeOutputStream out = null;

	private static File file = null;

	private static VideoFormat format = QuickTimeOutputStream.VideoFormat.JPG;

	private static TimerTask timerTask = null;
	private static Timer timer = null;

	private static int frameCount = 0;

	//CLI Params
	private static String url = null;
	private static String filename = null;
	private static Integer interval = null;
	private static int frames = -1;
	private static float quality = 1f;
	private static int fps = 30;
	private static int h = 0;
	private static int w = 0;
	private static int frameduration = 1;

	/**
	 * @param args
	 */
	public static void main(String[] args) {
		if(args.length < 1){
			System.out.println("Usage: WebCamTimeLapse URL [interval] [frames#|-1] [filename] [fps] [frameduration] [height] [width] [quality]");
			return;
		}

		//Mi registro per il CTRL+C
		Runtime.getRuntime().addShutdownHook(new Thread(new WebCamTimeLapse()));

		try {
			url = args[0];
			//if(!url.contains(".jpg") && url.contains(".jpeg")) throw new Exception("Not a jpg image in URL!");
			//Testo validita' URL e che all'indirizzo ci sia un'img valida
			img = ImageIO.read(new URL(url));
			if(img == null) throw new Exception("Not a jpg image in URL!");
		} catch (Exception e1) {
			System.out.println("Invalid URL. Exiting...");
			return;
		}
		try{
			interval = new Integer(args[1]);
		} catch (Exception e1) {
			interval = new Integer(120);
			System.out.println("Invalid/Missing capture interval. Using "+interval+" seconds");
		}
		try{
			frames = new Integer(args[2]);
		} catch (Exception e1) {
			frames = new Integer(-1).intValue();
			System.out.println("Invalid/Missing frames count. Using infinite. Stop using CTRL+C");
		}
		try{
			filename = args[3];
			//Se il file esiste lo rinomino
			if(new File(filename).exists()){
				System.out.println("Warning: "+filename+" already existing. Renaming to "+filename+"_old.mov");
				new File(filename).renameTo(new File(filename+"_old.mov"));
			}
			//Provo a vedere se posso scrivere
			new File(filename).canWrite();
		} catch (Exception e1) {
			if(url.contains("jpg")||url.contains("jpeg")) filename = url.substring(url.lastIndexOf("/") + 1, url.length()).replace(".jpg", "").replace(".jpeg", "")+".mov";
		    else filename = "Output_"+System.currentTimeMillis()+".mov";
			System.out.println("Invalid/Missing filename. Using "+filename);
			//Se il file esiste lo rinomino
			if(new File(filename).exists()){
				System.out.println("Warning: "+filename+" already existing. Renamed to "+filename.replace(".mov", "")+"_old.mov");
				new File(filename).renameTo(new File(filename+"_old.mov"));
			}
		}
		try{
			fps = new Integer(args[4]).intValue();
		}
		catch (Exception e){
			fps = 15;
			System.out.println("Invalid/Missing FPS. Using "+fps);
		}
		try{
			frameduration = new Integer(args[5]);
		} catch (Exception e1) {
			frameduration = 1;
			System.out.println("Invalid/Missing frame duration. Using "+(float)frameduration/(float)fps+" seconds");
		}
		try{
			h = new Integer(args[6]).intValue();
			w = new Integer(args[7]).intValue();
		}
		catch (Exception e){
			try{
				if(img == null) img = ImageIO.read(new URL(url));
				w = img.getWidth();
				h = img.getHeight();
				System.out.println("Invalid/Missing size. Using "+w+"x"+h);
			}
			catch (MalformedURLException e1) { }
			catch (IOException e1) { }
		}
		try{
			quality = new Float(args[8]).floatValue();
		}
		catch (Exception e){
			quality = 1f;
			System.out.println("Invalid quality. Using "+quality);
		}
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
				addFrame();
			}
		};

		synchronized (timerTask) {
			try {
				timer = new Timer();
				timer.scheduleAtFixedRate(timerTask, 0, interval.intValue()*1000);
				System.out.println("------------------------------------------------------");
				System.out.println("Input URL "+url);
				System.out.println("Output filename " +filename);
				System.out.println("Capture Interval "+interval+" seconds");
				if(frames > 0) System.out.println("Output frames " +frames);
				System.out.println("Output FPS "+fps+" (Frame duration " +(float)frameduration/(float)fps+" seconds)");
				if(frames >= 0)	{
					System.out.println("Output lenght "+((float)frames/(float)fps)+" seconds");
					System.out.println("Job completition time "+(frames*interval/60)+" minutes");
				}
				System.out.println("Output size "+w+"*"+h);
				System.out.println("Output quality "+quality);
				System.out.println("------------------------------------------------------");
				System.out.println(new Date().toString()+" Capture job started...");
				timerTask.wait();
			}
			catch (InterruptedException e){ }
		}

		timer.cancel();
		timer = null;
		timerTask.cancel();
		timerTask = null;

		if (out != null) {
			try {
				out.close();
			}
			catch (IOException e) {
				System.out.println("Error closing file!");
			}
		}
		System.out.println("Done!");
	}
	
	private static void addFrame(){
		try{
			img = watermark(ImageIO.read(new URL(url)),"WebCamTimeLapse");
			out.writeFrame(img, frameduration);
			frameCount++;
			System.out.print(new Date().toString() + " Frame "+frameCount+" added.");
			if(frames >= 0)	System.out.print(" Remaining "+(frames*interval/60)+" minutes.");
			System.out.println(" Current video duration "+((float)frameCount/(float)fps)+" seconds");
			synchronized (timerTask) {
				if (frameCount == frames)
					timerTask.notify();
			}
		}
		catch (MalformedURLException e) {}
		catch (IOException e) {
			System.out.print(new Date().toString() + " Error adding frame.");
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
	
	@Override
	public void run() {	
		timer.cancel();
		timer = null;
		timerTask.cancel();
		timerTask = null;

		if (out != null) {
			try {
				out.close();
			}
			catch (IOException e) {
				System.out.println("Error closing file!");
			}
		}
		System.out.println("Done!");
	}
}
